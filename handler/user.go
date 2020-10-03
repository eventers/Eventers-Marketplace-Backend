package handler

import (
	"encoding/json"
	"eventers-marketplace-backend/factory"
	"eventers-marketplace-backend/firebase"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/response"
	"eventers-marketplace-backend/user"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

const (
	a_provider  = "apple"
	fb_provider = "facebook"
	g_provider  = "google"
	p_provider  = "phone"
)

func CreateUser(service *user.User, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.CreateUser
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			response.BadRequest("invalid request body", fmt.Sprintf("createUser: error unmarshalling request body: %+v", err)).Send(ctx, w)
			return
		}

		if err := validate(req.Data.User); err != nil {
			response.InvalidData(fmt.Sprintf("createUser: invalid request: %+v", err)).Send(ctx, w)
			return
		}

		token, err := firebase.VerifyIDToken(ctx, f.FirebaseApp(ctx), req.Data.Auth.TokenID)
		if err != nil {
			response.Unauthorized().Send(ctx, w)
			return
		}

		if token.UID != firebaseID(req.Data.User) {
			response.FirebaseInvalidUID().Send(ctx, w)
			logger.Errorf(ctx, "create: unable to verify firebase user id: %v", firebaseID(req.Data.User))
			return
		}

		ipAddress, err := getIP(r)
		if ipAddress == "" || err != nil {
			logger.Infof(ctx, "create: unable to resolve ip address from header, continuing: %+v", err)
		}

		usr, auth, err := service.Create(ctx, f.DB(ctx), req.Data.User, req.Data.Auth, ipAddress)
		if err != nil {
			logger.Errorf(ctx, "create: unable to create user: %+v", err)
			response.SomethingWrong().Send(ctx, w)
		}

		response.SuccessResponse{
			Data: &response.Data{
				User: usr,
				Auth: auth,
			},
			StatusCode: http.StatusCreated,
		}.Send(w)
	}
}

func VerifyUser(service *user.User, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.CreateUser
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Errorf(ctx, "createUser: error unmarshalling request body: %+v", err)
			response.BadRequest("invalid request body", "").Send(ctx, w)
			return
		}

		if req.Data.User.UserID <= 0 ||
			strings.TrimSpace(*req.Data.User.PhoneCountryCode) == "" ||
			strings.TrimSpace(*req.Data.User.PhoneNumber) == "" {
			response.InvalidData("verifyUser: phone country code or number is invalid").Send(ctx, w)
		}

		//_, ok := firebase.VerifyJWTIDToken(viper.GetString(config.FirebaseProjectID), req.Data.Auth.TokenID, time.Duration(viper.GetInt(config.JWTOfflineInterval)))
		//if !ok {
		//	response.Unauthorized().Send(ctx, w)
		//	return
		//}

		ipAddress, err := getIP(r)
		if ipAddress == "" || err != nil {
			logger.Infof(ctx, "create: unable to resolve ip address from header, continuing: %+v", err)
		}

		user, res, err := service.Verify(ctx, f.DB(ctx), req.Data.User, req.Data.Auth, ipAddress)
		if err != nil || res != nil {
			response.SomethingWrong().Send(ctx, w)
			return
		}

		response.SuccessResponse{
			Data:       &response.Data{User: user},
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}
func validate(u *model.User) error {
	if u.Provider == nil {
		return fmt.Errorf("validate: no provider provided")
	}
	switch *u.Provider {
	case a_provider:
		if isEmpty(u.AFirebaseID) {
			return fmt.Errorf("validate: %s: invalid data", a_provider)
		}

		if !isEmpty(u.AEmail) && !validateEmail(*u.AEmail) {
			return fmt.Errorf("validate: %s: invalid email id", fb_provider)
		}
	case fb_provider:
		if isEmpty(u.FBFirebaseID) {
			return fmt.Errorf("validate: %s: invalid data", fb_provider)
		}

		if !isEmpty(u.FBEmail) && !validateEmail(*u.FBEmail) {
			return fmt.Errorf("validate: %s: invalid email id", fb_provider)
		}

	case g_provider:
		if isEmpty(u.GFirebaseID) {
			return fmt.Errorf("validate: %s: invalid data", g_provider)
		}

		if !isEmpty(u.GEmail) && !validateEmail(*u.GEmail) {
			return fmt.Errorf("validate: %s: invalid email id", g_provider)
		}

	case p_provider:
		if isEmpty(u.PhoneFirebaseID) || isEmpty(u.PhoneCountryCode) || isEmpty(u.PhoneNumber) {
			return fmt.Errorf("validate: %s: invalid data", p_provider)
		}

	default:
		return fmt.Errorf("validate: invalid provider")
	}

	return nil
}

func validateEmail(email string) bool {
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if len(email) > 254 || !rxEmail.MatchString(email) {
		return false
	}

	return true
}

func getIP(req *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("getIP: %q is not IP:port", req.RemoteAddr)
	}

	userIP := net.ParseIP(ip)

	if userIP != nil {
		return string(userIP), nil
	}
	forward := req.Header.Get("X-Forwarded-For")

	return strings.Split(forward, ",")[0], nil
}

func firebaseID(u *model.User) string {
	switch *u.Provider {

	case fb_provider:
		return *u.FBFirebaseID

	case g_provider:
		return *u.GFirebaseID

	default:
		return *u.PhoneFirebaseID
	}

}

func isEmpty(s *string) bool {
	return s == nil || strings.TrimSpace(*s) == ""
}
