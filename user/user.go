package user

import (
	"context"
	"database/sql"
	"errors"
	"eventers-marketplace-backend/algorand"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/response"
	"eventers-marketplace-backend/vault"
	"fmt"
	"strings"
	"time"
)

const (
	a_provider                 = "apple"
	fb_provider                = "facebook"
	g_provider                 = "google"
	p_provider                 = "phone"
	mobile_number_verification = "MOBILE_NUMBER_VERIFICATION"
)

var (
	a_columns      = []string{"a_firebase_id", "a_email", "a_name", "a_image_url"}
	f_columns      = []string{"fb_firebase_id", "fb_email", "fb_name", "fb_image_url"}
	g_columns      = []string{"g_firebase_id", "g_email", "g_name", "g_image_url"}
	p_columns      = []string{"phone_firebase_id", "phone_country_code", "phone_number"}
	ext_columns    = []string{"first_name", "last_name", "display_name", "email_address", "city", "state", "country", "address", "pincode"}
	date           = []string{"updated_date"}
	userDeviceCols = []string{"user_id", "device_id", "fcm_token", "ip_address", "login_time", "device_provider", "is_valid"}
	isRegistered   = "is_registered"
	isActive       = "is_active"
	NoRecordFound  = errors.New("no record found")
)

func NewUser(algo algorand.Algo, v vault.Vault) *User {
	return &User{Algo: algo, Vault: v}
}

type User struct {
	Algo  algorand.Algo
	Vault vault.Vault
}

// Get returns users profile
func (u *User) Get(db *sql.DB, userID int64, firebaseID string) (*model.User, error) {
	id, err := fetchUserID(db, firebaseID)
	if err != nil {
		return nil, fmt.Errorf("get: error getting user id from firebase id: %w", err)
	}

	if id != userID {
		return nil, fmt.Errorf("get: user id mismatch: %d:%d", userID, id)
	}
	user, err := fetchUser(db, userID)
	if err != nil {
		return nil, fmt.Errorf("get: error getting user profile: %w", err)
	}

	account, ok, err := u.userAddress(fmt.Sprintf("%v%v/0", user.PhoneCountryCode, user.PhoneNumber))
	if err != nil {
		return nil, fmt.Errorf("get: error getting user account: %w", err)
	}
	if !ok {
		return user, nil
	}

	user.AccountAddress = account.AccountAddress

	return user, nil
}

// Create creates a new user on database
func (u *User) Create(ctx context.Context, db *sql.DB, eu *model.User, a *model.Auth, ipAddress string) (*model.User, *model.Auth, error) {
	if db == nil || eu == nil || a == nil || eu.Provider == nil {
		return nil, nil, fmt.Errorf("create: invalid params")
	}

	if eu.Provider != nil && *eu.Provider == p_provider {
		return phoneProvider(ctx, db, u.Vault, u.Algo, eu, a, ipAddress)
	}

	return socialProvider(ctx, db, eu, a, ipAddress)
}

func (u *User) Verify(ctx context.Context, db *sql.DB, eu *model.User, auth *model.Auth, ipAddress string) (*model.User, *response.ErrorResponse, error) {

	cols := []string{"phone_country_code", "phone_number", "phone_firebase_id", "is_registered", "is_active"}
	vals := []interface{}{eu.PhoneCountryCode, eu.PhoneNumber, eu.PhoneFirebaseID, 1, 1}
	err := update(db, eu.UserID, cols, vals)
	if err != nil && !errors.Is(err, NoRecordFound) {
		res := response.SomethingWrong()
		return nil, &res, fmt.Errorf("verify: unable to update record: %s", err)
	}

	columns, values, err := provider(p_provider, eu)
	if err != nil {
		return nil, nil, fmt.Errorf("verify: unable to get phone details: %s", err)
	}

	columns = append(columns, "is_registered")
	columns = append(columns, "is_active")

	values = append(values, 1)
	values = append(values, 1)

	user, found, err := fetch(db, columns, values, eu.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("verify: unable to check firebase id: %s", err)
	}

	if found {
		cols := providerColumns(*eu.Provider)

		err = copy(db, append(cols, "phone_firebase_id"), eu.UserID, user.UserID)
		if err != nil {
			res := response.SomethingWrong()
			return nil, &res, fmt.Errorf("verify: unable to update records: %s", err)
		}

		//err = fcm(db, ipAddress, user.UserID, auth)
		//if err != nil {
		//	return nil, nil, fmt.Errorf("verify: unable to upsert user_device: %s", err)
		//}

		usr, err := fetchUser(db, user.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("verify: error fetching user: id: %d: err: %w", user.UserID, err)
		}
		return usr, nil, nil
	}

	//err = fcm(db, ipAddress, eu.UserID, auth)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("verify: unable to upsert user_device: %s", err)
	//}

	usr, err := fetchUser(db, eu.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("verify: error fetching user: id: %d: err: %w", eu.UserID, err)
	}

	err = saveAddress(ctx, u.Vault, u.Algo, fmt.Sprintf("%v%v/0", usr.PhoneCountryCode, usr.PhoneNumber))
	if err != nil {
		logger.Errorf(ctx, "verify: error saving address: %+v", err)
	}

	return usr, nil, nil
}

func (u *User) Update(db *sql.DB, user *model.User) (*model.User, error) {
	if user.UserID <= 0 {
		return nil, fmt.Errorf("update: invalid user_id provided: %d", user.UserID)
	}
	cols, vals := userColsVals(user)
	err := update(db, user.UserID, cols, vals)
	if err != nil {
		return nil, fmt.Errorf("update: error updating user details: %s", err)
	}
	return user, nil
}

func phoneProvider(ctx context.Context, db *sql.DB, v vault.Vault, algo algorand.Algo, user *model.User, a *model.Auth, ipAddress string) (*model.User, *model.Auth, error) {
	columns, values, err := provider(*user.Provider, user)
	if err != nil {
		return nil, nil, fmt.Errorf("create: unable to get provider: %s", err)
	}

	columns = append(columns, "is_registered")
	columns = append(columns, "is_active")
	values = append(values, 1)
	values = append(values, 1)
	u, found, err := Exists(db, columns, values)
	if err != nil {
		return nil, nil, fmt.Errorf("create: unable to check firebase id: %s", err)
	}

	if found && u.IsRegistered && u.IsActive {
		if u.PhoneFirebaseID != user.PhoneFirebaseID {
			logger.Infof(ctx, "phoneProvider: case 1: old id: %s: new id: %s", u.PhoneFirebaseID, user.PhoneFirebaseID)
		}

		cols, _, args := columnsAndValues(*user)
		err = update(db, u.UserID, cols, args)
		if err != nil {
			return nil, nil, fmt.Errorf("phoneProvider: case 1: unable to update details: %s", err)
		}

		//err = fcm(db, ipAddress, u.UserID, a)
		//if err != nil {
		//	return nil, nil, fmt.Errorf("phoneProvider: unable to update user_device: %s", err)
		//}

		usr, err := fetchUser(db, u.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("phoneProvider: error fetching user: id: %d: err: %w", u.UserID, err)
		}
		return usr, nil, nil
	}

	n := len(values)
	values[n-2] = 0
	values[n-1] = 0
	columns = append(columns, "fb_firebase_id")
	columns = append(columns, "g_firebase_id")
	values = append(values, nil)
	values = append(values, nil)
	u, found, err = Exists(db, columns, values)
	if err != nil {
		return nil, nil, fmt.Errorf("phoneProvider: unable to check firebase id: %s", err)
	}
	if found {
		if u.PhoneFirebaseID != user.PhoneFirebaseID {
			logger.Infof(ctx, "phoneProvider: case 1: old id: %s: new id: %s", u.PhoneFirebaseID, user.PhoneFirebaseID)
		}

		cols, _, args := columnsAndValues(*user)
		err = update(db, u.UserID, cols, args)
		if err != nil {
			return nil, nil, fmt.Errorf("phoneProvider: case 2:unable to update details: %s", err)
		}

		//err = fcm(db, ipAddress, u.UserID, a)
		//if err != nil {
		//	return nil, nil, fmt.Errorf("phoneProvider: unable to update user_device: %s", err)
		//}

		usr, err := fetchUser(db, u.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("phoneProvider: error fetching user: id: %d: err: %w", u.UserID, err)
		}

		err = saveAddress(ctx, v, algo, fmt.Sprintf("%v%v/0", usr.PhoneCountryCode, usr.PhoneNumber))
		if err != nil {
			logger.Errorf(ctx, "phoneProvider: error saving address: %+v", err)
		}
		return usr, nil, nil
	}

	cols, params, args := columnsAndValues(*user)

	query := fmt.Sprintf(`INSERT INTO Users (%s) VALUES (%s);`, strings.Join(cols, ", "), strings.Join(params, ", "))

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, nil, fmt.Errorf("phoneProvider: unable to prepare query: %s", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)

	id, err := result.LastInsertId()
	if err != nil {
		return nil, nil, fmt.Errorf("phoneProvider: unable to get last insert id: %s", err)
	}

	//err = fcm(db, ipAddress, u.UserID, a)
	//if err != nil {
	//	return nil, nil, fmt.Errorf("phoneProvider: unable to update user_device: %s", err)
	//}

	usr, err := fetchUser(db, id)
	if err != nil {
		return nil, nil, fmt.Errorf("phoneProvider: error fetching user: id: %d: err: %w", id, err)
	}

	err = saveAddress(ctx, v, algo, fmt.Sprintf("%v%v/0", usr.PhoneCountryCode, usr.PhoneNumber))
	if err != nil {
		logger.Errorf(ctx, "phoneProvider: error saving address: %+v", err)
	}
	return usr, nil, nil

}

func socialProvider(ctx context.Context, db *sql.DB, user *model.User, a *model.Auth, ipAddress string) (*model.User, *model.Auth, error) {
	columns, values, err := provider(*user.Provider, user)
	if err != nil {
		return nil, nil, fmt.Errorf("socialProvider: unable to get provider: %s", err)
	}

	columns = append(columns, "is_registered")
	columns = append(columns, "is_active")
	values = append(values, 1)
	values = append(values, 1)
	u, found, err := Exists(db, columns, values)
	if err != nil {
		return nil, nil, fmt.Errorf("socialProvider: unable to check firebase id: %s", err)
	}

	if found && u.IsRegistered && u.IsActive {
		cols, _, args := columnsAndValues(*user)
		err = update(db, u.UserID, cols, args)
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: case 1: unable to update details: %s", err)
		}

		//err = fcm(db, ipAddress, u.UserID, a)
		//if err != nil {
		//	return nil, nil, fmt.Errorf("socialProvider: unable to update user_device: %s", err)
		//}

		usr, err := fetchUser(db, u.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: error fetching user: id: %d: err: %w", u.UserID, err)
		}
		return usr, nil, nil
	}

	n := len(values)
	values[n-2] = 0
	values[n-1] = 0
	u, found, err = Exists(db, columns, values)
	if err != nil {
		return nil, nil, fmt.Errorf("socialProvider: unable to check firebase id: %s", err)
	}
	// Case 2: *_firebase_id Exists and is_registered = 0

	if found && !u.IsRegistered && !u.IsActive {
		phoneCountryCode, phoneNumber, err := retrieveMobile(db, u.UserID)
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: case 2: unable to retrieve mobile details: %s", err)
		}

		user := &model.User{UserID: u.UserID, PhoneCountryCode: &phoneCountryCode, PhoneNumber: &phoneNumber}
		auth := &model.Auth{Status: mobile_number_verification}
		return user, auth, nil
	}

	columns = columns[:n-2]
	values = values[:n-2]
	u, found, err = Exists(db, columns, values)
	if err != nil {
		return nil, nil, fmt.Errorf("socialProvider: unable to check firebase id: %s", err)
	}

	// Case 3 and 4: *_firebase_id does not exist and *_email != null and *_email Exists in !*_email
	var email *string
	var cols []interface{}
	var emails []interface{}
	switch *user.Provider {
	case a_provider:
		cols = []interface{}{"fb_email", "g_email"}
		emails = []interface{}{user.AEmail, user.AEmail}
		email = user.AEmail
	case fb_provider:
		cols = []interface{}{"a_email", "g_email"}
		emails = []interface{}{user.FBEmail, user.FBEmail}
		email = user.FBEmail
	default:
		cols = []interface{}{"a_email", "fb_email"}
		emails = []interface{}{user.GEmail, user.GEmail}
		email = user.GEmail
	}

	if !found && !isEmpty(email) {

		// Case 3
		var u *model.User
		var found bool
		var err error
		var col interface{}
		var email interface{}
		for i := range cols {
			col = cols[i]
			email = emails[i]
			u, found, err = Exists(db, []interface{}{col, "is_registered", "is_active"}, []interface{}{email, 0, 0})
			if err != nil {
				return nil, nil, fmt.Errorf("socialProvider: case 3: unable to check emai id: %w", err)
			}
			if found {
				break
			}
		}

		if found {
			cols, _, args := columnsAndValues(*user)
			err = update(db, u.UserID, cols, args)
			if err != nil {
				return nil, nil, fmt.Errorf("socialProvider: case 3: unable to update details: %s", err)
			}

			phoneCountryCode, phoneNumber, err := retrieveMobile(db, u.UserID)
			if err != nil {
				return nil, nil, fmt.Errorf("socialProvider: case 3: unable to retrieve mobile details: %s", err)
			}

			user := &model.User{UserID: u.UserID, PhoneCountryCode: &phoneCountryCode, PhoneNumber: &phoneNumber}
			auth := &model.Auth{Status: mobile_number_verification}
			return user, auth, nil
		}

		u, found, err = Exists(db, []interface{}{col, "is_registered", "is_active"}, []interface{}{email, 1, 1})
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: case 4: unable to check firebase id: %s", err)
		}

		if found {
			cols, _, args := columnsAndValues(*user)
			err = update(db, u.UserID, cols, args)
			if err != nil {
				return nil, nil, fmt.Errorf("socialProvider: case 3: unable to update details: %s", err)
			}

			//err = fcm(db, ipAddress, u.UserID, a)
			//if err != nil {
			//	return nil, nil, fmt.Errorf("socialProvider: unable to update user_device: %s", err)
			//}

			usr, err := fetchUser(db, u.UserID)
			if err != nil {
				return nil, nil, fmt.Errorf("socialProvider: error fetching user: id: %d: err: %w", u.UserID, err)
			}
			return usr, nil, nil
		}

	}

	if !found {
		cols, params, args := columnsAndValues(*user)

		query := fmt.Sprintf(`INSERT INTO Users (%s) VALUES (%s);`, strings.Join(cols, ", "), strings.Join(params, ", "))

		stmt, err := db.Prepare(query)
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: unable to prepare query: %s", err)
		}
		defer stmt.Close()

		result, err := stmt.Exec(args...)

		id, err := result.LastInsertId()
		if err != nil {
			return nil, nil, fmt.Errorf("socialProvider: unable to get last insert id: %s", err)
		}

		eu := &model.User{UserID: id}
		auth := &model.Auth{Status: mobile_number_verification}
		return eu, auth, nil
	}

	return nil, nil, fmt.Errorf("socialProvider: error")
}

//func fcm(db *sql.DB, ipAdrress string, userID int64, a *model.Auth) error {
//	query := `SELECT user_device_id FROM User_Device WHERE user_id = ? AND device_id = ?`
//
//	stmt, err := db.Prepare(query)
//	if err != nil {
//		return fmt.Errorf("fcm: error preparing query: %s", err)
//	}
//	defer stmt.Close()
//
//	rows, err := stmt.Query(userID, a.DeviceID)
//	if err != nil {
//		return fmt.Errorf("fcm: error executing search query: %s", err)
//	}
//
//	var userDeviceID int64
//	if rows.Next() {
//		err := rows.Scan(&userDeviceID)
//		if err != nil {
//			return fmt.Errorf("fcm: error scanning user_device_id: %s", err)
//		}
//	}
//	defer rows.Close()
//
//	if userDeviceID > 0 {
//		query := fmt.Sprintf(`UPDATE User_Device SET fcm_token = ?, ip_address = ?, login_time = ?,
//						device_provider = ?, is_valid = ? WHERE user_device_id = ?`)
//
//		stmt, err := db.Prepare(query)
//		if err != nil {
//			return fmt.Errorf("fcm: error preparing update query: %s", err)
//		}
//		defer stmt.Close()
//
//		_, err = stmt.Exec(a.FCMToken, ipAdrress, time.Now(), a.DeviceProvider, 1, userDeviceID)
//		if err != nil {
//			return fmt.Errorf("fcm: error executing update query: %s", err)
//		}
//	} else {
//		var params []string
//		for _, _ = range userDeviceCols {
//			params = append(params, "?")
//		}
//		query := fmt.Sprintf(`INSERT INTO User_Device (%s) VALUES (%s);`, strings.Join(userDeviceCols, ", "), strings.Join(params, ", "))
//		stmt, err := db.Prepare(query)
//		if err != nil {
//			return fmt.Errorf("fcm: error preparing insert query: %s", err)
//		}
//		defer stmt.Close()
//
//		_, err = stmt.Exec(userID, a.DeviceID, a.FCMToken, ipAdrress, time.Now(), a.DeviceProvider, 1)
//		if err != nil {
//			return fmt.Errorf("fcm: error executing insert query: %s", err)
//		}
//	}
//
//	return nil
//}

func fetch(db *sql.DB, columns, values []interface{}, existingUserID int64) (*model.User, bool, error) {
	var withPlaceholder []string
	for _, col := range columns {
		withPlaceholder = append(withPlaceholder, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf(
		`SELECT user_id, is_registered, is_active, phone_firebase_id, a_firebase_id, fb_firebase_id, 
				g_firebase_id, a_email, fb_email, g_email FROM Users WHERE %s  and user_id <> ?`,
		strings.Join(withPlaceholder, " and "),
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("fetc: error preparing query for %#v: %s", columns, err)
	}
	defer stmt.Close()

	values = append(values, existingUserID)
	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, false, fmt.Errorf("fetch: error executing query for %#v: %s", columns, err)
	}

	var u model.User
	if rows.Next() {
		err := rows.Scan(
			&u.UserID,
			&u.IsRegistered,
			&u.IsActive,
			&u.PhoneFirebaseID,
			&u.AFirebaseID,
			&u.FBFirebaseID,
			&u.GFirebaseID,
			&u.AEmail,
			&u.FBEmail,
			&u.GEmail,
		)
		if err != nil {
			return nil, false, fmt.Errorf("fetch: error while scanning row: %s", err)
		}
		return &u, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}
func copy(db *sql.DB, columns []string, from, to int64) error {
	var cols []string

	for _, v := range columns {
		cols = append(cols, fmt.Sprintf("u1.%s = u2.%s", v, v))
	}

	query := fmt.Sprintf(`
					UPDATE Users u1
					INNER JOIN Users u2
					ON u2.phone_country_code = u1.phone_country_code
					AND u2.phone_number 	 = u1.phone_number
					SET %s , u2.is_registered = 1, u2.is_active = 0
					WHERE u1.user_id = ? and u2.user_id = ?
				`, strings.Join(cols, ", "))

	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("copy: error preparing query to copy: %s", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(to, from)
	if err != nil {
		return fmt.Errorf("copy: error executing query: %s", err)
	}

	return nil
}

func update(db *sql.DB, userID int64, cols []string, values []interface{}) error {
	values = append(values, userID)
	var set []string

	for _, col := range cols {
		set = append(set, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf(`UPDATE Users SET %s WHERE user_id = ?;`, strings.Join(set, ","))

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(values...)
	if err != nil {
		return fmt.Errorf("update: unable to execute query: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update: unable to get rows affected: %s", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("update: %d rows affected: err: %w", rowsAffected, NoRecordFound)
	}

	return nil
}

func retrieveMobile(db *sql.DB, userID int64) (string, string, error) {
	query := "SELECT phone_country_code, phone_number FROM Users WHERE user_id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", "", fmt.Errorf("Exists: error preparing query for phone number: %s", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(userID)

	var phoneCountryCode sql.NullString
	var phoneNumber sql.NullString
	if row != nil {
		row.Scan(&phoneCountryCode, &phoneNumber)
	}

	return phoneCountryCode.String, phoneNumber.String, nil
}

func Exists(db *sql.DB, columns, values []interface{}) (*model.User, bool, error) {
	var withPlaceholder []string
	for _, col := range columns {
		withPlaceholder = append(withPlaceholder, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf(
		`SELECT user_id, first_name, last_name, display_name, email_address, city, state, country, address, pincode,
			fb_firebase_id, fb_email, fb_name, fb_image_url, g_firebase_id, g_email, g_name, g_image_url, 
			a_firebase_id, a_email, a_name, a_image_url, phone_firebase_id, phone_country_code, phone_number, 
			provider, profile_pic, is_registered, is_active FROM Users WHERE %s`,
		strings.Join(withPlaceholder, " and "),
	)

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, false, fmt.Errorf("Exists: error preparing query for %#v: %s", columns, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(values...)
	if err != nil {
		return nil, false, fmt.Errorf("Exists: error executing query for %#v: %s", columns, err)
	}

	var u model.User
	if rows.Next() {
		err := rows.Scan(
			&u.UserID,
			&u.FirstName,
			&u.LastName,
			&u.DisplayName,
			&u.EmailAddress,
			&u.City,
			&u.State,
			&u.Country,
			&u.Address,
			&u.Pincode,
			&u.FBFirebaseID,
			&u.FBEmail,
			&u.FBName,
			&u.FBImageURL,
			&u.GFirebaseID,
			&u.GEmail,
			&u.GName,
			&u.GImageURL,
			&u.AFirebaseID,
			&u.AEmail,
			&u.AName,
			&u.AImageURL,
			&u.PhoneFirebaseID,
			&u.PhoneCountryCode,
			&u.PhoneNumber,
			&u.Provider,
			&u.ProfilePic,
			&u.IsRegistered,
			&u.IsActive,
		)
		if err != nil {
			return nil, false, fmt.Errorf("Exists: error while scanning row: %s", err)
		}
		return &u, true, nil
	}
	defer rows.Close()

	return nil, false, nil
}

func provider(provider string, u *model.User) ([]interface{}, []interface{}, error) {
	switch provider {

	case a_provider:
		return []interface{}{a_columns[0]}, []interface{}{u.AFirebaseID}, nil

	case fb_provider:
		return []interface{}{f_columns[0]}, []interface{}{u.FBFirebaseID}, nil

	case g_provider:
		return []interface{}{g_columns[0]}, []interface{}{u.GFirebaseID}, nil

	case p_provider:
		return []interface{}{p_columns[1], p_columns[2]}, []interface{}{u.PhoneCountryCode, u.PhoneNumber}, nil
	}

	return nil, nil, fmt.Errorf("provider: invalid provider")
}

func columnsAndValues(u model.User) ([]string, []string, []interface{}) {
	columns := []string{"provider"}

	var args []interface{}

	switch *u.Provider {

	case a_provider:
		columns = append(columns, a_columns...)
		args = append(args, []interface{}{u.AFirebaseID, u.AEmail, u.AName, u.AImageURL}...)

	case fb_provider:
		columns = append(columns, f_columns...)
		args = append(args, []interface{}{u.FBFirebaseID, u.FBEmail, u.FBName, u.FBImageURL}...)

	case g_provider:
		columns = append(columns, g_columns...)
		args = append(args, []interface{}{u.GFirebaseID, u.GEmail, u.GName, u.GImageURL}...)

	default:
		columns = append(columns, p_columns...)
		columns = append(columns, isRegistered)
		columns = append(columns, isActive)
		args = append(args, []interface{}{u.PhoneFirebaseID, u.PhoneCountryCode, u.PhoneNumber, true, true}...)
	}

	columns = append(columns, date...)

	args = append([]interface{}{
		u.Provider,
	}, args...)

	args = append(args, time.Now())

	var values []string
	for _, _ = range columns {
		values = append(values, "?")
	}

	return columns, values, args
}

func providerColumns(provider string) []string {

	switch provider {
	case a_provider:
		return a_columns
	case fb_provider:
		return f_columns
	case g_provider:
		return g_columns
	default:
		return p_columns
	}
}

func userColsVals(u *model.User) ([]string, []interface{}) {
	var vals []interface{}
	var cols []string

	if !isEmpty(u.FirstName) {
		cols = append(cols, "first_name")
		vals = append(vals, u.FirstName)
	}

	if !isEmpty(u.LastName) {
		cols = append(cols, "last_name")
		vals = append(vals, u.LastName)
	}

	if !isEmpty(u.DisplayName) {
		cols = append(cols, "display_name")
		vals = append(vals, u.DisplayName)
	}

	if !isEmpty(u.EmailAddress) {
		cols = append(cols, "email_name")
		vals = append(vals, u.EmailAddress)
	}

	if !isEmpty(u.City) {
		cols = append(cols, "city")
		vals = append(vals, u.City)
	}

	if !isEmpty(u.State) {
		cols = append(cols, "state")
		vals = append(vals, u.State)
	}

	if !isEmpty(u.Country) {
		cols = append(cols, "country")
		vals = append(vals, u.Country)
	}

	if !isEmpty(u.Address) {
		cols = append(cols, "address")
		vals = append(vals, u.Address)
	}

	if !isEmpty(u.Pincode) {
		cols = append(cols, "pincode")
		vals = append(vals, u.Pincode)
	}

	if !isEmpty(u.ProfilePic) {
		cols = append(cols, "profile_pic")
		vals = append(vals, u.ProfilePic)
	}

	return cols, vals
}

func isEmpty(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == ""
}

func fetchUser(db *sql.DB, userID int64) (*model.User, error) {
	user := model.User{UserID: userID}
	query := `SELECT first_name, last_name, display_name, email_address, city, state, country, address, pincode,
			fb_firebase_id, fb_email, fb_name, fb_image_url, g_firebase_id, g_email, g_name, g_image_url, 
			a_firebase_id, a_email, a_name, a_image_url, phone_firebase_id, phone_country_code, phone_number, 
			provider, profile_pic from Users where user_id=?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("fetchUser: error preparing query to fetch user for the id: %d: %w", userID, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, fmt.Errorf("fetchUser: error executing query to fetch user for the id: %d: %w", userID, err)
	}

	if rows.Next() {
		err := rows.Scan(
			&user.FirstName,
			&user.LastName,
			&user.DisplayName,
			&user.EmailAddress,
			&user.City,
			&user.State,
			&user.Country,
			&user.Address,
			&user.Pincode,
			&user.FBFirebaseID,
			&user.FBEmail,
			&user.FBName,
			&user.FBImageURL,
			&user.GFirebaseID,
			&user.GEmail,
			&user.GName,
			&user.GImageURL,
			&user.AFirebaseID,
			&user.AEmail,
			&user.AName,
			&user.AImageURL,
			&user.PhoneFirebaseID,
			&user.PhoneCountryCode,
			&user.PhoneNumber,
			&user.Provider,
			&user.ProfilePic,
		)
		if err != nil {
			return nil, fmt.Errorf("fetchUser: error while scanning row: %s", err)
		}
	}
	defer rows.Close()

	return &user, nil
}

func fetchUserID(db *sql.DB, firebaseID string) (int64, error) {
	var userID int64
	query := "SELECT user_id FROM Users WHERE (a_firebase_id=? OR fb_firebase_id=? OR g_firebase_id=? OR phone_firebase_id=?) AND is_active = 1 AND is_registered = 1;"
	stmt, err := db.Prepare(query)
	if err != nil {
		return userID, fmt.Errorf("fetchUserID: error preparing query to fetch user id: %s: %w", firebaseID, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(firebaseID, firebaseID, firebaseID, firebaseID)
	if err != nil {
		return userID, fmt.Errorf("fetchUserID: error executing query to fetch user for the id: %s: %w", firebaseID, err)
	}

	if rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			return userID, fmt.Errorf("fetchUserID: error while scanning row: %s", err)
		}
	}
	defer rows.Close()

	return userID, nil
}
