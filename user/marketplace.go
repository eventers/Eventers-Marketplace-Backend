package user

import (
	"context"
	"database/sql"
	"eventers-marketplace-backend/codec"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/response"
	"eventers-marketplace-backend/twilio"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/pquerna/otp/totp"
)

const (
	interop    = "INTEROP"
	noninterop = "NON_INTEROP"
	otpMessage = "OTP to verify your number at eventers is: %s"
	otp_sent   = "OTP_SUCCESSFULLY_SENT"
)

func (u *User) CreateMarketPlaceUser(ctx context.Context, db *sql.DB, eu *model.MarketplaceUser, a *model.Auth, sender twilio.Sender, client *redis.Client, secret string) (*model.MarketplaceUser, *model.Auth, error) {

	m, ok, err := MarketPlaceExists(db, []interface{}{"access_key"}, []interface{}{a.AccessKey})
	if err != nil {
		return nil, nil, response.SomethingWrong()
	}
	if !ok {
		return nil, nil, response.Unauthorized()
	}

	user, ok, err := MarketPlaceUserExists(db, []interface{}{"user_phone_number", "marketplace_id"}, []interface{}{formatPhoneNumber(eu), m.MarketPlaceID})
	if err != nil {
		return nil, nil, response.SomethingWrong()
	}

	if !ok {

		query := `INSERT INTO User_Marketplace (marketplace_id, user_phone_number) VALUES (?, ?);`

		// Execute query
		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, nil, fmt.Errorf("CreateMarketPlaceUser: unable to prepare query: %s", err)
		}
		defer stmt.Close()

		result, err := stmt.Exec(m.MarketPlaceID, formatPhoneNumber(eu))
		if err != nil {
			logger.Errorf(ctx, "CreateMarketPlaceUser: unable to get last insert id: %+v", err)
			return nil, nil, response.SomethingWrong()
		}

		// Get user_id of the user
		id, err := result.LastInsertId()
		if err != nil {
			logger.Errorf(ctx, "CreateMarketPlaceUser: unable to get last insert id: %+v", err)
			return nil, nil, response.SomethingWrong()
		}

		eu.UserID = id

		if sendOTP(sender, client, id, secret, formatPhoneNumber(eu)) != nil {
			logger.Errorf(ctx, "CreateMarketPlaceUser: error sending otp: %+v", err)
			return nil, nil, response.SomethingWrong()
		}
		return eu, &model.Auth{Status: otp_sent}, nil
	}

	eu.UserID = user.UserID
	if user.IsValid && m.AccessType == interop {
		return eu, nil, nil
	}

	if sendOTP(sender, client, user.UserID, secret, formatPhoneNumber(eu)) != nil {
		logger.Errorf(ctx, "CreateMarketPlaceUser: error sending otp: %+v", err)
		return nil, nil, response.SomethingWrong()
	}

	return eu, &model.Auth{Status: otp_sent}, nil
}

func sendOTP(sender twilio.Sender, client *redis.Client, userMarketplaceID int64, secret, phoneNumber string) error {
	otp, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("sendOTP: unable to generate otp: %s", err)
	}

	sid, err := sender.Send(phoneNumber, fmt.Sprintf(otpMessage, otp))
	if err != nil {
		return fmt.Errorf("sendOTP: unable to send otp to: %s: %s", phoneNumber, err)
	}

	err = client.Set(fmt.Sprintf("marketplace-%d", userMarketplaceID), otp, time.Minute*5).Err()
	if err != nil {
		return fmt.Errorf("sendOTP: unable to save otp into the db for mobile: %s, sid: %v : %s", phoneNumber, sid, err)
	}

	return nil
}

func (u *User) VerifyMarketPlaceUserOTP(ctx context.Context, db *sql.DB, client *redis.Client, eu *model.MarketplaceUser, auth model.Auth) (*model.MarketplaceUser, error) {
	m, ok, err := MarketPlaceExists(db, []interface{}{"access_key"}, []interface{}{auth.AccessKey})
	if err != nil {
		return nil, response.SomethingWrong()
	}
	if !ok {
		return nil, response.Unauthorized()
	}

	user, ok, err := MarketPlaceUserExists(db, []interface{}{"user_phone_number", "marketplace_id"}, []interface{}{formatPhoneNumber(eu), m.MarketPlaceID})
	if err != nil {
		return nil, response.SomethingWrong()
	}

	if !ok {
		return nil, response.UserNotExist()
	}

	key := client.Get(fmt.Sprintf("marketplace-%d", eu.UserID))
	if key.Err() != nil {
		return nil, response.OTPExpired()
	}

	if key.Val() != auth.OTP {
		return nil, response.OTPMismatch()
	}

	path := fmt.Sprintf("%s", formatPhoneNumber(eu))
	if m.AccessType == interop {
		path = path + "/0"
	} else {
		path = fmt.Sprintf("%s/%v", path, m.MarketPlaceID)
	}

	if user.IsValid {
		if m.AccessType == noninterop {
			err = u.encryptKeys(eu, m, path)
			if err != nil {
				logger.Errorf(ctx, "verifyMarketPlaceUserOTP: could no encrypt private key: %+v", err)
				return nil, response.SomethingWrong()
			}
		}
		return eu, nil
	}

	err = saveAddress(ctx, u.Vault, u.Algo, path)
	if err != nil {
		logger.Errorf(ctx, "unable to save private key to vault: %+v", err)
		return nil, response.SomethingWrong()
	}

	query := `UPDATE User_Marketplace SET is_valid = ? WHERE user_marketplace_id = ?;`

	stmt, err := db.Prepare(query)
	if err != nil {
		logger.Errorf(ctx, "VerifyMarketPlaceUserOTP: error preparing update query: %+v", err)
		return nil, response.SomethingWrong()
	}
	defer stmt.Close()

	result, err := stmt.Exec(true, eu.UserID)
	if err != nil {
		logger.Errorf(ctx, "VerifyMarketPlaceUserOTP: unable to execute query: %+v", err)
		return nil, response.SomethingWrong()
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Errorf(ctx, "VerifyMarketPlaceUserOTP: unable to get rows affected: %s", err)
		return nil, response.SomethingWrong()
	}

	if rowsAffected != 1 {
		logger.Errorf(ctx, "update: %d rows affected: err: %+v", rowsAffected, NoRecordFound)
		return nil, response.SomethingWrong()
	}
	if m.AccessType == noninterop {
		err = u.encryptKeys(eu, m, path)
		if err != nil {
			logger.Errorf(ctx, "verifyMarketPlaceUserOTP: could no encrypt private key: %+v", err)
			return nil, response.SomethingWrong()
		}
	}

	return eu, nil
}

func (u *User) encryptKeys(eu *model.MarketplaceUser, m *model.Marketplace, path string) error {
	account, ok, err := u.userAddress(path)
	if err != nil {
		return fmt.Errorf("encryptKey: error fetching user address: %w", err)
	}
	if !ok {
		return fmt.Errorf("encryptKey: user account does not exist on vault: %w", err)
	}

	encryptedAddress, err := codec.Encrypt([]byte(m.AccessKey), []byte( account.AccountAddress))
	if err != nil {
		return fmt.Errorf("encryptKey: could no encrypt account address: %w", err)
	}

	encryptedPassphrase, err := codec.Encrypt([]byte(m.AccessKey), []byte( account.SecurityPassphrase))
	if err != nil {
		return fmt.Errorf("encryptKey: could no encrypt passphrase: %w", err)
	}
	eu.AccountAddress = encryptedAddress
	eu.AccountPassphrase = encryptedPassphrase
	return nil
}

func MarketPlaceExists(db *sql.DB, columns, values []interface{}) (*model.Marketplace, bool, error) {
	var withPlaceholder []string
	for _, col := range columns {
		withPlaceholder = append(withPlaceholder, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf(
		`SELECT marketplace_id, marketplace_name, access_key, access_type  FROM Marketplace WHERE %s`,
		strings.Join(withPlaceholder, " and "),
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("MarketPlaceExists: error preparing query for %#v: %s", columns, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, false, fmt.Errorf("MarketPlaceExists: error executing query for %#v: %s", columns, err)
	}

	var u model.Marketplace
	if rows.Next() {
		err := rows.Scan(
			&u.MarketPlaceID,
			&u.MarketPlaceName,
			&u.AccessKey,
			&u.AccessType,
		)
		if err != nil {
			return nil, false, fmt.Errorf("MarketPlaceExists: error while scanning row: %s", err)
		}
		return &u, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}

func MarketPlaceUserExists(db *sql.DB, columns, values []interface{}) (*model.MarketplaceUser, bool, error) {
	var withPlaceholder []string
	for _, col := range columns {
		withPlaceholder = append(withPlaceholder, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf(
		`SELECT user_marketplace_id, marketplace_id, is_valid FROM User_Marketplace WHERE %s`,
		strings.Join(withPlaceholder, " and "),
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("MarketPlaceUserExists: error preparing query for %#v: %s", columns, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, false, fmt.Errorf("MarketPlaceUserExists: error executing query for %#v: %s", columns, err)
	}

	var u model.MarketplaceUser
	if rows.Next() {
		err := rows.Scan(
			&u.UserID,
			&u.MarketPlaceID,
			&u.IsValid,
		)
		if err != nil {
			return nil, false, fmt.Errorf("MarketPlaceUserExists: error while scanning row: %s", err)
		}
		return &u, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}

func formatPhoneNumber(u *model.MarketplaceUser) string {
	return fmt.Sprintf("%s%s", u.PhoneCountryCode, u.PhoneNumber)
}
