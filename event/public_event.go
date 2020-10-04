package event

import (
	"context"
	"database/sql"
	"eventers-marketplace-backend/algorand"
	"eventers-marketplace-backend/constants"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/vault"
	"fmt"
	"strings"
)

const (
	publicEventTable = "Public_Event"
	eventTicketTable = "Event_Tickets"
	seedAlgo         = 10
)

var active = "ACTIVE"

var publicEventCols = []string{"date_time", "event_title", "event_description", "event_image", "total_tickets", "ticket_price", "temp_account_address", "temp_security_paraphrase"}
var eventTicketCols = []string{"business_user_id", "public_event_id", "asset_id", "current_holder_id", "status", "price"}

// NewEvent returns a new event database instance
func NewEvent(algo algorand.Algo, vault vault.Vault) *Event {
	return &Event{
		algo:  algo,
		vault: vault,
	}
}

// Event represents the client for event table
type Event struct {
	algo  algorand.Algo
	vault vault.Vault
}

func (u *Event) PublicEvent(ctx context.Context, db *sql.DB, pe *model.PublicEvent, addedBy int64) (*model.PublicEvent, error) {

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("publicEvent: error begining db transaction: %s", err)
	}

	a, err := u.algo.GenerateAccount()
	if err != nil {
		return nil, fmt.Errorf("publicEvent: error generating account: %w", err)
	}

	values := []interface{}{
		pe.DateTime,
		pe.EventTitle,
		pe.EventDescription,
		pe.EventImage,
		pe.TotalTickets,
		pe.TicketPrice,
		a.AccountAddress,
		a.SecurityPassphrase,
	}

	id, err := create(tx, publicEventTable, publicEventCols, values)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("publicEvent: error inserting public events by: %d: err: %w", addedBy, err)
	}

	pe.PublicEventID = id

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("publicEvent: error commiting public event to db by: %d: err: %s", addedBy, err)
	}

	path := fmt.Sprintf("%s/%v", u.vault.TempPath, pe.PublicEventID)
	data := map[string]interface{}{
		constants.AccountAddress:     a.AccountAddress,
		constants.PrivateKey:         a.PrivateKey,
		constants.SecurityPassphrase: a.SecurityPassphrase,
	}
	_, err = u.vault.Logical().Write(path, data)
	if err != nil {
		return nil, fmt.Errorf("PublicEvent: unable to write to vault: %w", err)
	}

	go u.processEvent(ctx, db, a, pe, addedBy)
	return pe, nil
}

// Update Event_Tickets
func (u *Event) UpdatePublicEvent(ctx context.Context, db *sql.DB, et *model.Ticket) error {
	if et.PriceToResell > 0 {
		err := u.resell(db, et)
		if err != nil {
			return fmt.Errorf("updatePublicEvent: error in reselling: %w", err)
		}
		return nil
	}

	if et.FromUserID > 0 && et.ToUserID > 0 {
		err := u.send(ctx, db, et)
		if err != nil {
			return fmt.Errorf("updatePublicEvent: error in sending: %w", err)
		}
		return nil
	}

	if et.Status != nil && *et.Status == "REDEEM" {
		err := u.redeem(db, et)
		if err != nil {
			return fmt.Errorf("updatePublicEvent: error in redeeming: %w", err)
		}
		return nil
	}

	if et.PublicEventID > 0 && et.EventTicketID == 0 {
		err := u.buy(ctx, db, et)
		if err != nil {
			return fmt.Errorf("updatePublicEvent: error in buying: %w", err)
		}
		return nil
	}

	if et.EventTicketID > 0 && et.PublicEventID > 0 {
		err := u.buyResell(ctx, db, et)
		if err != nil {
			return fmt.Errorf("updatePublicEvent: error in buyresell: %w", err)
		}
		return nil
	}

	return fmt.Errorf("updatePublicEvent: no matching action found")
}

func (u *Event) resell(db *sql.DB, et *model.Ticket) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("resell: error begining db transaction: %s", err)
	}
	updatedRows, err := update(
		tx,
		eventTicketTable,
		[]string{"price", "status"},
		[]interface{}{et.PriceToResell, "RESELL"},
		[]string{"event_ticket_id"},
		[]interface{}{et.EventTicketID},
	)

	if err != nil {
		return fmt.Errorf("resell: error updating event_ticket for resell: %w", err)
	}

	if updatedRows == 0 {
		tx.Rollback()
		return fmt.Errorf("resell: no row updated")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("resell: could not commit transaction for resell: err: %w", err)
	}

	return nil
}

func (u *Event) redeem(db *sql.DB, et *model.Ticket) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("resell: error begining db transaction: %s", err)
	}
	updatedRows, err := update(
		tx,
		eventTicketTable,
		[]string{"status"},
		[]interface{}{"REDEEM"},
		[]string{"event_ticket_id"},
		[]interface{}{et.EventTicketID},
	)

	if err != nil {
		return fmt.Errorf("redeem: error updating event_ticket for redeem: %w", err)
	}

	if updatedRows == 0 {
		tx.Rollback()
		return fmt.Errorf("redeem: no row updated")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("redeem: could not commit transaction for redeem: err: %w", err)
	}

	return nil
}

func (u *Event) send(ctx context.Context, db *sql.DB, et *model.Ticket) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("send: error begining db transaction: %s", err)
	}
	to, ok, err := u.fetchUserAddress(et.ToUserID)
	if err != nil {
		return fmt.Errorf("send: error fetching to_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("send: to_user_id not found")
	}

	from, ok, err := u.fetchUserAddress(et.FromUserID)
	if err != nil {
		return fmt.Errorf("send: error fetching from_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("send: from_user_id not found")
	}

	eventTicket, ok, err := fetchEventTicket(db, et.EventTicketID)
	if err != nil {
		return fmt.Errorf("send: error fetching event ticket: %w", err)
	}

	if !ok {
		return fmt.Errorf("send: event_ticket_id not found")
	}

	toAccount := algorand.Account{
		AccountAddress:     to.AccountAddress,
		PrivateKey:         to.PrivateKey,
		SecurityPassphrase: to.SecurityPassphrase,
	}

	fromAccount := algorand.Account{
		AccountAddress:     from.AccountAddress,
		PrivateKey:         from.PrivateKey,
		SecurityPassphrase: from.SecurityPassphrase,
	}

	err = u.algo.OptIn(ctx, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("send: error opting in: %w", err)
	}

	err = u.algo.SendAsset(ctx, &fromAccount, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("send: error sending asset: %w", err)
	}

	updatedRows, err := update(
		tx,
		eventTicketTable,
		[]string{"current_holder_id"},
		[]interface{}{et.ToUserID},
		[]string{"event_ticket_id"},
		[]interface{}{et.EventTicketID},
	)

	if err != nil {
		return fmt.Errorf("send: error updating event_ticket for send: %w", err)
	}

	if updatedRows == 0 {
		tx.Rollback()
		return fmt.Errorf("send: no row updated")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("resell: could not commit transaction for resell: err: %w", err)
	}

	return nil
}

func (u *Event) buy(ctx context.Context, db *sql.DB, et *model.Ticket) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("buy: error begining db transaction: %s", err)
	}

	to, ok, err := u.fetchUserAddress(et.ToUserID)
	if err != nil {
		return fmt.Errorf("buy: error fetching to_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("buy: to_user_id not found")
	}

	eventTicket, ok, err := pickEventTicket(db, et.PublicEventID)
	if err != nil {
		return fmt.Errorf("buy: error picking event ticket: %w", err)
	}

	if !ok {
		return fmt.Errorf("buy: no active ticket found")
	}

	from, ok, err := u.fetchUserAddress(eventTicket.BusinessUserID)
	if err != nil {
		return fmt.Errorf("buy: error fetching from_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("buy: from_user_id not found")
	}

	toAccount := algorand.Account{
		AccountAddress:     to.AccountAddress,
		PrivateKey:         to.PrivateKey,
		SecurityPassphrase: to.SecurityPassphrase,
	}

	fromAccount := algorand.Account{
		AccountAddress:     from.AccountAddress,
		PrivateKey:         from.PrivateKey,
		SecurityPassphrase: from.SecurityPassphrase,
	}

	err = u.algo.OptIn(ctx, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("buy: error opting in: %w", err)
	}

	err = u.algo.SendAsset(ctx, &fromAccount, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("buy: error buying asset: %w", err)
	}

	updatedRows, err := update(
		tx,
		eventTicketTable,
		[]string{"current_holder_id"},
		[]interface{}{et.ToUserID},
		[]string{"event_ticket_id"},
		[]interface{}{eventTicket.EventTicketID},
	)

	if err != nil {
		return fmt.Errorf("buy: error updating event_ticket for buy: %w", err)
	}

	if updatedRows == 0 {
		tx.Rollback()
		return fmt.Errorf("buy: no row updated")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("buy: could not commit transaction for resell: err: %w", err)
	}

	return nil
}

func (u *Event) buyResell(ctx context.Context, db *sql.DB, et *model.Ticket) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("buyResell: error begining db transaction: %s", err)
	}

	to, ok, err := u.fetchUserAddress(et.ToUserID)
	if err != nil {
		return fmt.Errorf("buyResell: error fetching to_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("buyResell: to_user_id not found")
	}

	eventTicket, ok, err := fetchEventTicket(db, et.EventTicketID)
	if err != nil {
		return fmt.Errorf("buyResell: error fetching event ticket: %w", err)
	}

	if !ok {
		return fmt.Errorf("buyResell: no active ticket found")
	}

	from, ok, err := u.fetchUserAddress(eventTicket.BusinessUserID)
	if err != nil {
		return fmt.Errorf("buyResell: error fetching from_user_id: %w", err)
	}

	if !ok {
		return fmt.Errorf("buyResell: from_user_id not found")
	}

	toAccount := algorand.Account{
		AccountAddress:     to.AccountAddress,
		PrivateKey:         to.PrivateKey,
		SecurityPassphrase: to.SecurityPassphrase,
	}

	fromAccount := algorand.Account{
		AccountAddress:     from.AccountAddress,
		PrivateKey:         from.PrivateKey,
		SecurityPassphrase: from.SecurityPassphrase,
	}

	err = u.algo.OptIn(ctx, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("buyResell: error opting in: %w", err)
	}

	err = u.algo.SendAsset(ctx, &fromAccount, &toAccount, eventTicket.AssetID)
	if err != nil {
		return fmt.Errorf("buyResell: error buying asset: %w", err)
	}

	updatedRows, err := update(
		tx,
		eventTicketTable,
		[]string{"current_holder_id", "status"},
		[]interface{}{et.ToUserID, active},
		[]string{"event_ticket_id"},
		[]interface{}{et.EventTicketID},
	)

	if err != nil {
		return fmt.Errorf("buyResell: error updating event_ticket for buyResell: %w", err)
	}

	if updatedRows == 0 {
		tx.Rollback()
		return fmt.Errorf("buyResell: no row updated")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("buyResell: could not commit transaction for resell: err: %w", err)
	}

	return nil
}

func (u *Event) processEvent(ctx context.Context, db *sql.DB, a *algorand.Account, pe *model.PublicEvent, userID int64) {
	err := u.algo.Send(ctx, a, seedAlgo)
	if err != nil {
		logger.Errorf(ctx, "processEvent: error sending 10 algos to temp account: %s: err: %+v", a.AccountAddress, err)
		return
	}

	ua, ok, err := u.fetchUserAddress(userID)
	if err != nil {
		logger.Errorf(ctx, "processEvent: could not fetchUserAddress user, err: %+v", err)
		return
	}

	if !ok {
		logger.Errorf(ctx, "processEvent: user not found")
		return
	}

	var et []model.EventTicket
	var assets []uint64
	var i uint64 = 0
	for ; i < pe.TotalTickets; i++ {
		assetID, err := u.algo.CreateAsset(ctx, a)
		if err != nil {
			logger.Errorf(ctx, "processEvent: error creating asset: %+v", err)
		}
		eventTicket := model.EventTicket{
			BusinessUserID:  userID,
			PublicEventID:   pe.PublicEventID,
			AssetID:         assetID,
			CurrentHolderID: userID,
			Status:          &active,
			Price:           pe.TicketPrice,
		}

		assets = append(assets, assetID)
		et = append(et, eventTicket)

		go u.start(ctx, db, &eventTicket, a, *ua, assetID, userID)
	}

}

func (u *Event) start(ctx context.Context, db *sql.DB, et *model.EventTicket, from *algorand.Account, ua algorand.Account, assetID uint64, userID int64) {
	ac := algorand.Account{
		AccountAddress:     ua.AccountAddress,
		PrivateKey:         ua.PrivateKey,
		SecurityPassphrase: ua.SecurityPassphrase,
	}

	err := u.algo.OptIn(ctx, &ac, assetID)
	if err != nil {
		logger.Errorf(ctx, "start: error opting in for the asset: ID: %d, err: %+v", assetID, err)
		return
	}

	err = u.algo.SendAsset(ctx, from, &ac, assetID)
	if err != nil {
		logger.Errorf(ctx, "start: could not transfer asset: ID: %d, err: %+v", assetID, err)
		return
	}

	err = u.createTicket(db, et, userID)
	if err != nil {
		logger.Errorf(ctx, "start: could not insert ticket: ID: %d, err: %+v", assetID, err)
	}
}

func (u *Event) createTicket(db *sql.DB, event *model.EventTicket, sharedBy int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("createTicket: error begining db transaction: %s", err)
	}

	values := []interface{}{
		event.BusinessUserID,
		event.PublicEventID,
		event.AssetID,
		event.CurrentHolderID,
		event.Status,
		event.Price,
	}

	id, err := create(tx, eventTicketTable, eventTicketCols, values)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("createTicket: error inserting event ticket: err: %w", err)
	}
	event.EventTicketID = id

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("createTicket: error commiting event tickets:err: %w", err)
	}

	return nil
}

func (u *Event) fetchUserAddress(userID int64) (*algorand.Account, bool, error) {
	path := fmt.Sprintf("%s/%v", u.vault.UserPath, userID)
	secret, err := u.vault.Logical().Read(path)
	if err != nil {
		return nil, false, fmt.Errorf("fetchUserAddress: could not fetchUserAddress of user: %d", userID)
	}

	accountAddress, accountAddressOK := secret.Data[constants.AccountAddress]
	if !accountAddressOK {
		return nil, false, fmt.Errorf("fetchUserAddress: account address not found")
	}
	privateKey, privateKeyOK := secret.Data[constants.PrivateKey]
	if !privateKeyOK {
		return nil, false, fmt.Errorf("fetchUserAddress: private key not found")
	}
	securityPassphrase, securityPassphraseOK := secret.Data[constants.SecurityPassphrase]
	if !securityPassphraseOK {
		return nil, false, fmt.Errorf("fetchUserAddress: security passphrase not found")
	}

	ua := algorand.Account{
		AccountAddress:     accountAddress.(string),
		PrivateKey:         privateKey.(string),
		SecurityPassphrase: securityPassphrase.(string),
	}

	return &ua, true, nil
}

func fetchEventTicket(db *sql.DB, eventTicketID int64) (*model.EventTicket, bool, error) {
	query := fmt.Sprintf(
		`SELECT event_ticket_id, business_user_id, public_event_id, asset_id, current_holder_id, status,
				available_to_resell, price FROM Event_Tickets WHERE event_ticket_id = ?`,
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("fetchEventTicket: error preparing query: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(eventTicketID)
	if err != nil {
		return nil, false, fmt.Errorf("fetchEventTicket: error executing query: %w", err)
	}

	var ua model.EventTicket
	if rows.Next() {
		err := rows.Scan(
			&ua.EventTicketID,
			&ua.BusinessUserID,
			&ua.PublicEventID,
			&ua.AssetID,
			&ua.CurrentHolderID,
			&ua.Status,
			&ua.AvailableToResell,
			&ua.Price,
		)
		if err != nil {
			return nil, false, fmt.Errorf("fetchEventTicket: error while scanning row: %s", err)
		}
		return &ua, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}

func pickEventTicket(db *sql.DB, publicEventID int64) (*model.EventTicket, bool, error) {
	query := fmt.Sprintf(
		`SELECT event_ticket_id, business_user_id, public_event_id, asset_id, current_holder_id, status,
				available_to_resell, price FROM Event_Tickets 
				WHERE business_user_id = current_holder_id AND status=? AND public_event_id = ?`,
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("pickEventTicket: error preparing query: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(active, publicEventID)
	if err != nil {
		return nil, false, fmt.Errorf("pickEventTicket: error executing query: %w", err)
	}

	var ua model.EventTicket
	if rows.Next() {
		err := rows.Scan(
			&ua.EventTicketID,
			&ua.BusinessUserID,
			&ua.PublicEventID,
			&ua.AssetID,
			&ua.CurrentHolderID,
			&ua.Status,
			&ua.AvailableToResell,
			&ua.Price,
		)
		if err != nil {
			return nil, false, fmt.Errorf("pickEventTicket: error while scanning row: %s", err)
		}
		return &ua, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}

func create(tx *sql.Tx, table string, cols []string, values []interface{}) (int64, error) {
	var params []string

	for range cols {
		params = append(params, "?")
	}

	tsql := fmt.Sprintf(`INSERT INTO %s(%s) VALUES (%s);`, table, strings.Join(cols, ", "), strings.Join(params, ", "))

	// Execute query
	stmt, err := tx.Prepare(tsql)
	if err != nil {
		return -1, fmt.Errorf("create: error preparing sql query: %s", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)

	if err != nil {
		return -1, fmt.Errorf("create: unable to insert record in %s: %s", table, err)
	}

	return result.LastInsertId()

}

func update(tx *sql.Tx, table string, cols []string, values []interface{}, column []string, value []interface{}) (int64, error) {
	values = append(values, value...)
	var set []string

	for _, col := range cols {
		set = append(set, fmt.Sprintf("%s = ?", col))
	}

	var params []string

	for range cols {
		params = append(params, "?")
	}

	var conds []string

	for _, c := range column {
		conds = append(conds, fmt.Sprintf("%s = ?", c))
	}

	tsql := fmt.Sprintf(`UPDATE %s SET %s WHERE  %s;`, table, strings.Join(set, ","), strings.Join(conds, " AND "))

	// Execute query
	stmt, err := tx.Prepare(tsql)
	if err != nil {
		return -1, fmt.Errorf("update: error preparing sql query: %s", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)

	if err != nil {
		return -1, fmt.Errorf("update: unable to insert record in %s: %s", table, err)
	}

	return result.RowsAffected()
}
