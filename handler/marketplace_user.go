package handler

import (
	"encoding/json"
	"eventers-marketplace-backend/factory"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/response"
	"eventers-marketplace-backend/twilio"
	"eventers-marketplace-backend/user"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
)

func CreateMarketPlaceUser(service *user.User, f factory.Factory, sender twilio.Sender, client *redis.Client, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.CreateMarketPlaceUser
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			response.BadRequest("invalid request body", fmt.Sprintf("createMarketPlaceUser: error unmarshalling request body: %+v", err)).Send(ctx, w)
			return
		}

		usr, auth, err := service.CreateMarketPlaceUser(ctx, f.DB(ctx), req.Data.User, req.Data.Auth, sender, client, secret)
		if err != nil {
			logger.Errorf(ctx, "createMarketPlaceUser: unable to create user: %+v", err)
			if e, ok := err.(response.ErrorResponse); ok {
				e.Send(ctx, w)
				return
			}
			response.SomethingWrong().Send(ctx, w)
		}

		response.SuccessResponse{
			Data: &response.Data{
				UserMarketplace: usr,
				Auth:            auth,
			},
			StatusCode: http.StatusCreated,
		}.Send(w)
	}
}

func VerifyMarketPlaceOTP(service *user.User, f factory.Factory, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.CreateMarketPlaceUser
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			response.BadRequest("invalid request body", fmt.Sprintf("verifyMarketPlaceOTP: error unmarshalling request body: %+v", err)).Send(ctx, w)
			return
		}

		user, err := service.VerifyMarketPlaceUserOTP(ctx, f.DB(ctx), client, req.Data.User, *req.Data.Auth)
		if err != nil {
			if e, ok := err.(response.ErrorResponse); ok {
				e.Send(ctx, w)
				return
			}
			response.SomethingWrong().Send(ctx, w)
			return
		}

		response.SuccessResponse{
			Data:       &response.Data{UserMarketplace: user},
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}
