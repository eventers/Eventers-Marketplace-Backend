package algorand

import (
	"context"
	"encoding/base64"
	"eventers-marketplace-backend/logger"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/algorand/go-algorand-sdk/transaction"
)

type Algo interface {
	GenerateAccount() (*Account, error)
	Send(context.Context, *Account, uint64) error
	CreateAsset(context.Context, *Account) (uint64, error)
	OptIn(context.Context, *Account, uint64) error
	SendAsset(context.Context, *Account, *Account, uint64) error
}

type algo struct {
	from         *Account
	apiAddress   string
	apiKey       string
	amountFactor uint64
	minFee       uint64
	seedAlgo     uint64
}

func New(from *Account, apiAddress, apiKey string, amountFactor, minFee, seedAlgo uint64) Algo {
	return &algo{
		from:         from,
		apiAddress:   apiAddress,
		apiKey:       apiKey,
		amountFactor: amountFactor,
		minFee:       minFee,
		seedAlgo:     seedAlgo,
	}
}

func (a *algo) Send(ctx context.Context, to *Account, noOfAlgos uint64) error {
	var headers []*algod.Header
	headers = append(headers, &algod.Header{Key: "X-API-Key", Value: a.apiKey})
	algodClient, err := algod.MakeClientWithHeaders(a.apiAddress, "", headers)
	if err != nil {
		return fmt.Errorf("send: error connecting to algo: %w", err)
	}

	txParams, err := algodClient.SuggestedParams()
	if err != nil {
		return fmt.Errorf("send: error getting suggested tx params: %w", err)
	}

	fromAddr := a.from.AccountAddress
	toAddr := to.AccountAddress
	amount := noOfAlgos * a.amountFactor
	note := []byte(fmt.Sprintf("Transferring %d algos from %s", a.seedAlgo, a.from))
	genID := txParams.GenesisID
	genHash := txParams.GenesisHash
	firstValidRound := txParams.LastRound
	lastValidRound := firstValidRound + 1000

	txn, err := transaction.MakePaymentTxnWithFlatFee(fromAddr, toAddr, a.minFee, amount, firstValidRound, lastValidRound, note, "", genID, genHash)
	if err != nil {
		return fmt.Errorf("send: error creating transaction: %w", err)
	}

	privateKey, err := mnemonic.ToPrivateKey(a.from.SecurityPassphrase)
	if err != nil {
		return fmt.Errorf("send: error getting private key from mnemonic: %w", err)
	}

	txId, bytes, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		return fmt.Errorf("send: failed to sign transaction: %w", err)
	}
	logger.Infof(ctx, "Signed txid: %s", txId)

	txHeaders := append([]*algod.Header{}, &algod.Header{Key: "Content-Type", Value: "application/x-binary"})
	sendResponse, err := algodClient.SendRawTransaction(bytes, txHeaders...)
	if err != nil {
		return fmt.Errorf("send: failed to send transaction: %w", err)
	}
	logger.Infof(ctx, "send: submitted transaction %s", sendResponse.TxID)

	return nil
}

func (a *algo) GenerateAccount() (*Account, error) {
	account := crypto.GenerateAccount()
	paraphrase, err := mnemonic.FromPrivateKey(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("generateAccount: error generating account: %w", err)
	}

	return &Account{
		AccountAddress:     account.Address.String(),
		PrivateKey:         string(account.PrivateKey),
		SecurityPassphrase: paraphrase,
	}, nil
}

func (a *algo) CreateAsset(ctx context.Context, ac *Account) (uint64, error) {

	var headers []*algod.Header
	headers = append(headers, &algod.Header{Key: "X-API-Key", Value: a.apiKey})
	algodClient, err := algod.MakeClientWithHeaders(a.apiAddress, "", headers)
	if err != nil {
		return 0, fmt.Errorf("createAsset: error connecting to algo: %w", err)
	}

	txParams, err := algodClient.SuggestedParams()
	if err != nil {
		return 0, fmt.Errorf("createAsset: error getting suggested tx params: %w", err)
	}

	genID := txParams.GenesisID
	genHash := txParams.GenesisHash
	firstValidRound := txParams.LastRound
	lastValidRound := firstValidRound + 1000

	// Create an asset
	// Set parameters for asset creation transaction
	creator := ac.AccountAddress
	assetName := "eventers"
	unitName := "tickets"
	assetURL := "https://www.eventersapp.com"
	assetMetadataHash := "thisIsSomeLength32HashCommitment"
	defaultFrozen := false
	decimals := uint32(0)
	totalIssuance := uint64(1)
	manager := a.from.AccountAddress
	reserve := a.from.AccountAddress
	freeze := ""
	clawback := a.from.AccountAddress
	note := []byte(nil)
	txn, err := transaction.MakeAssetCreateTxn(creator, a.minFee, firstValidRound, lastValidRound, note,
		genID, base64.StdEncoding.EncodeToString(genHash), totalIssuance, decimals, defaultFrozen, manager, reserve, freeze, clawback,
		unitName, assetName, assetURL, assetMetadataHash)
	if err != nil {
		return 0, fmt.Errorf("createAsset: failed to make asset: %w", err)
	}
	fmt.Printf("Asset created AssetName: %s\n", txn.AssetConfigTxnFields.AssetParams.AssetName)

	privateKey, err := mnemonic.ToPrivateKey(ac.SecurityPassphrase)
	if err != nil {
		return 0, fmt.Errorf("createAsset: error getting private key from mnemonic: %w", err)
	}

	txid, stx, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		return 0, fmt.Errorf("createAsset: failed to sign transaction: %w", err)
	}
	logger.Infof(ctx, "Signed txid: %s", txid)
	// Broadcast the transaction to the network
	txHeaders := append([]*algod.Header{}, &algod.Header{Key: "Content-Type", Value: "application/x-binary"})
	sendResponse, err := algodClient.SendRawTransaction(stx, txHeaders...)
	if err != nil {
		return 0, fmt.Errorf("createAsset: failed to send transaction: %w", err)
	}

	// Wait for transaction to be confirmed
	waitForConfirmation(ctx, algodClient, sendResponse.TxID)

	// Retrieve asset ID by grabbing the max asset ID
	// from the creator account's holdings.
	act, err := algodClient.AccountInformation(ac.AccountAddress)
	if err != nil {
		return 0, fmt.Errorf("createAsset: failed to get account information: %w", err)
	}

	assetID := uint64(0)
	for i := range act.AssetParams {
		if i > assetID {
			assetID = i
		}
	}

	logger.Infof(ctx, "createAsset: asset ID from AssetParams: %d", assetID)
	// Retrieve asset info.
	assetInfo, err := algodClient.AssetInformation(assetID)
	if err != nil {
		return 0, fmt.Errorf("createAsset: error getting asset info: %w", err)
	}

	logger.Infof(ctx, "createAsset: assets info: %+v", assetInfo)

	return assetID, nil
}

func (a *algo) OptIn(ctx context.Context, ac *Account, assetID uint64) error {
	var headers []*algod.Header
	headers = append(headers, &algod.Header{Key: "X-API-Key", Value: a.apiKey})
	algodClient, err := algod.MakeClientWithHeaders(a.apiAddress, "", headers)
	if err != nil {
		return fmt.Errorf("optin: error connecting to algo: %w", err)
	}

	txParams, err := algodClient.SuggestedParams()
	if err != nil {
		return fmt.Errorf("optin: error getting suggested tx params: %w", err)
	}

	note := []byte(fmt.Sprintf("Opting in from %s", ac.AccountAddress))
	genID := txParams.GenesisID
	genHash := txParams.GenesisHash
	firstValidRound := txParams.LastRound
	lastValidRound := firstValidRound + 1000

	// Account 3 opts in to receive asset
	txn, err := transaction.MakeAssetAcceptanceTxn(ac.AccountAddress, a.minFee, firstValidRound,
		lastValidRound, note, genID, base64.StdEncoding.EncodeToString(genHash), assetID)
	if err != nil {
		return fmt.Errorf("optin: failed to send transaction MakeAssetAcceptanceTxn: %w", err)
	}

	privateKey, err := mnemonic.ToPrivateKey(ac.SecurityPassphrase)
	if err != nil {
		return fmt.Errorf("optin: error getting private key from mnemonic: %w", err)
	}

	txid, stx, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		return fmt.Errorf("optin: failed to sign transaction: %w", err)
	}

	fmt.Printf("Transaction ID: %s\n", txid)
	// Broadcast the transaction to the network
	txHeaders := append([]*algod.Header{}, &algod.Header{Key: "Content-Type", Value: "application/x-binary"})
	sendResponse, err := algodClient.SendRawTransaction(stx, txHeaders...)
	if err != nil {
		return fmt.Errorf("optin: failed to send transaction: %w", err)
	}

	logger.Infof(ctx, "optin: transaction ID raw: %s", sendResponse.TxID)

	// Wait for transaction to be confirmed
	waitForConfirmation(ctx, algodClient, sendResponse.TxID)

	act, err := algodClient.AccountInformation(ac.AccountAddress)
	if err != nil {
		return fmt.Errorf("optin: failed to get account information: %w", err)
	}

	logger.Infof(ctx, "optin: account info: %+v", act.Assets[assetID])

	return nil
}

func (a *algo) SendAsset(ctx context.Context, from, to *Account, assetID uint64) error {
	var headers []*algod.Header
	headers = append(headers, &algod.Header{Key: "X-API-Key", Value: a.apiKey})
	algodClient, err := algod.MakeClientWithHeaders(a.apiAddress, "", headers)
	if err != nil {
		return fmt.Errorf("sendAsset: error connecting to algo: %w", err)
	}

	txParams, err := algodClient.SuggestedParams()
	if err != nil {
		return fmt.Errorf("sendAsset: error getting suggested tx params: %w", err)
	}

	note := []byte("Transferring asset")
	genID := txParams.GenesisID
	genHash := txParams.GenesisHash
	firstValidRound := txParams.LastRound
	lastValidRound := firstValidRound + 1000

	// Send  1 of asset from Account to Account
	sender := from.AccountAddress
	recipient := to.AccountAddress
	amount := uint64(1)
	closeRemainderTo := ""
	txn, err := transaction.MakeAssetTransferTxn(sender, recipient,
		closeRemainderTo, amount, a.minFee, firstValidRound, lastValidRound, note,
		genID, base64.StdEncoding.EncodeToString(genHash), assetID)
	if err != nil {
		return fmt.Errorf("sendAsset: failed to send transaction MakeAssetTransfer Txn: %w", err)
	}

	privateKey, err := mnemonic.ToPrivateKey(from.SecurityPassphrase)
	if err != nil {
		return fmt.Errorf("sendAsset: error getting private key from mnemonic: %w", err)
	}

	txid, stx, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		return fmt.Errorf("sendAsset: failed to sign transaction: %w", err)
	}
	fmt.Printf("Transaction ID: %s", txid)
	// Broadcast the transaction to the network
	txHeaders := append([]*algod.Header{}, &algod.Header{Key: "Content-Type", Value: "application/x-binary"})
	sendResponse, err := algodClient.SendRawTransaction(stx, txHeaders...)
	if err != nil {
		return fmt.Errorf("sendAsset: failed to send transaction: %w", err)
	}
	fmt.Printf("Transaction ID raw: %s\n", sendResponse.TxID)

	// Wait for transaction to be confirmed
	waitForConfirmation(ctx, algodClient, sendResponse.TxID)

	act, err := algodClient.AccountInformation(to.AccountAddress)
	if err != nil {
		return fmt.Errorf("sendAsset: failed to get account information: %w", err)
	}

	logger.Infof(ctx, "sendAsset: account info: %v", act.Assets[assetID])
	return nil
}

// Function that waits for a given txId to be confirmed by the network
func waitForConfirmation(ctx context.Context, algodClient algod.Client, txID string) {
	for {
		pt, err := algodClient.PendingTransactionInformation(txID)
		if err != nil {
			//fmt.Printf("waiting for confirmation... (pool error, if any): %s\n", err)
			logger.Infof(ctx, "waiting for confirmation... (pool error, if any): %s\n", err)
			continue
		}
		if pt.ConfirmedRound > 0 {
			//fmt.Printf("Transaction "+pt.TxID+" confirmed in round %d\n", pt.ConfirmedRound)
			logger.Infof(ctx, "Transaction "+pt.TxID+" confirmed in round %d\n", pt.ConfirmedRound)
			break
		}
		nodeStatus, err := algodClient.Status()
		if err != nil {
			//fmt.Printf("error getting algod status: %s\n", err)
			logger.Warnf(ctx, "error getting algod status: %s\n", err)
			return
		}
		algodClient.StatusAfterBlock(nodeStatus.LastRound + 1)
	}
}
