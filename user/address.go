package user

import (
	"context"
	"eventers-marketplace-backend/algorand"
	"eventers-marketplace-backend/constants"
	"eventers-marketplace-backend/vault"
	"fmt"
)

func saveAddress(ctx context.Context, v vault.Vault, algo algorand.Algo, userID int64) error {
	a, err := algo.GenerateAccount()
	if err != nil {
		return fmt.Errorf("saveAddress: error generating address: %w", err)
	}

	path := fmt.Sprintf("%s/%v", v.UserPath, userID)
	data := map[string]interface{}{
		constants.AccountAddress:     a.AccountAddress,
		constants.PrivateKey:         a.PrivateKey,
		constants.SecurityPassphrase: a.SecurityPassphrase,
	}
	_, err = v.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("saveAddress: unable to write to vault: %w", err)
	}

	err = algo.Send(ctx, a, 5)
	if err != nil {
		return fmt.Errorf("saveAddress: error sending algos to: %+v: err: %w", a, err)
	}

	return nil
}

func (u *User) userAddress(userID int64) (*algorand.Account, bool, error) {
	path := fmt.Sprintf("%s/%v", u.Vault.UserPath, userID)
	secret, err := u.Vault.Logical().Read(path)
	if err != nil {
		return nil, false, fmt.Errorf("userAddress: could not get account of user: %d", userID)
	}

	accountAddress, accountAddressOK := secret.Data[constants.AccountAddress]
	if !accountAddressOK {
		return nil, false, fmt.Errorf("userAddress: account address not found")
	}
	privateKey, privateKeyOK := secret.Data[constants.PrivateKey]
	if !privateKeyOK {
		return nil, false, fmt.Errorf("userAddress: private key not found")
	}
	securityPassphrase, securityPassphraseOK := secret.Data[constants.SecurityPassphrase]
	if !securityPassphraseOK {
		return nil, false, fmt.Errorf("userAddress: security passphrase not found")
	}

	ua := algorand.Account{
		AccountAddress:     accountAddress.(string),
		PrivateKey:         privateKey.(string),
		SecurityPassphrase: securityPassphrase.(string),
	}

	return &ua, true, nil
}
