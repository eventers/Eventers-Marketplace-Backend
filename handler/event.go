package handler

import (
	"encoding/json"
	"eventers-marketplace-backend/config"
	"eventers-marketplace-backend/event"
	"eventers-marketplace-backend/factory"
	"eventers-marketplace-backend/firebase"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/model"
	"eventers-marketplace-backend/response"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func PublicEvent(service *event.Event, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.PublicEventRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Errorf(ctx, "createUser: error unmarshalling request body: %+v", err)
			response.BadRequest("invalid request body", "").Send(ctx, w)
			return
		}

		_, ok := firebase.VerifyJWTIDToken(viper.GetString(config.FirebaseProjectID), req.Data.Auth.TokenID, time.Duration(viper.GetInt(config.JWTOfflineInterval)))
		if !ok {
			response.Unauthorized().Send(ctx, w)
			return
		}

		publicEvent, err := service.PublicEvent(ctx, f.DB(ctx), req.Data.PublicEvent, req.Data.Auth.UserID)

		if err != nil {
			response.SomethingWrong().Send(ctx, w)
			logger.Errorf(ctx, "publicEvent: unable to create public event: %w", err)
			return
		}

		auth := &model.Auth{PushKey: req.Data.Auth.PushKey}
		response.SuccessResponse{
			Data:       &response.Data{PublicEvent: publicEvent, Auth: auth},
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}

func UpdatePublicEvent(service *event.Event, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.PublicEventUpdateReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Errorf(ctx, "createUser: error unmarshalling request body: %+v", err)
			response.BadRequest("invalid request body", "").Send(ctx, w)
			return
		}

		_, ok := firebase.VerifyJWTIDToken(viper.GetString(config.FirebaseProjectID), req.Data.Auth.TokenID, time.Duration(viper.GetInt(config.JWTOfflineInterval)))
		if !ok {
			response.Unauthorized().Send(ctx, w)
			return
		}

		err = service.UpdatePublicEvent(ctx, f.DB(ctx), req.Data.Ticket)

		if err != nil {
			response.SomethingWrong().Send(ctx, w)
			logger.Errorf(ctx, "updatePublicEvent: unable to update public event: %w", err)
			return
		}

		auth := &model.Auth{PushKey: req.Data.Auth.PushKey}
		response.SuccessResponse{
			Data:       &response.Data{Auth: auth},
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}

func GetPublicEvents(service *event.Event, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req model.PublicEventUpdateReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			logger.Errorf(ctx, "createUser: error unmarshalling request body: %+v", err)
			response.BadRequest("invalid request body", "").Send(ctx, w)
			return
		}

		_, ok := firebase.VerifyJWTIDToken(viper.GetString(config.FirebaseProjectID), req.Data.Auth.TokenID, time.Duration(viper.GetInt(config.JWTOfflineInterval)))
		if !ok {
			response.Unauthorized().Send(ctx, w)
			return
		}

		publicEvents, err := service.GetPublicEvents(f.DB(ctx))

		if err != nil {
			response.SomethingWrong().Send(ctx, w)
			logger.Errorf(ctx, "getPublicEvents: unable to get public events: %w", err)
			return
		}

		response.SuccessResponse{
			Data:       publicEvents,
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}

func GetPublicEvent(service *event.Event, f factory.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		userIDString := vars["userID"]

		userID, err := strconv.ParseInt(userIDString, 10, 64)
		if err != nil {
			response.InvalidData(fmt.Sprintf("getPublicEvent: invalid user id: %v", userIDString))
			logger.Errorf(ctx, "getPublicEvent: unable to parse userID: %s: %w", userIDString, err)
			return
		}

		publicEvents, err := service.GetPublicEvent(f.DB(ctx), userID)

		if err != nil {
			response.SomethingWrong().Send(ctx, w)
			logger.Errorf(ctx, "getPublicEvent: unable to get public event: %w", err)
			return
		}

		response.SuccessResponse{
			Data:       publicEvents,
			StatusCode: http.StatusOK,
		}.Send(w)
	}
}
