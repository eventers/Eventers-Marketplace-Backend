package router

import (
	"context"
	"eventers-marketplace-backend/algorand"
	"eventers-marketplace-backend/config"
	"eventers-marketplace-backend/event"
	"eventers-marketplace-backend/factory"
	"eventers-marketplace-backend/handler"
	"eventers-marketplace-backend/healthcheck"
	"eventers-marketplace-backend/logger"
	"eventers-marketplace-backend/middleware"
	"eventers-marketplace-backend/response"
	"eventers-marketplace-backend/twilio"
	"eventers-marketplace-backend/user"
	"eventers-marketplace-backend/vault"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// Router returns the router for all the API handler.
func Router(ctx context.Context) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.SetCorrelationIDHeader)
	r.Use(middleware.PanicHandler)
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		response.ResourceNotFound(fmt.Sprintf("The requested resource was not found: path: %s, method: %s", req.URL.Path, req.Method), "The requested resource was not found!").Send(req.Context(), w)
	})

	r.Use(middleware.ResponseTimeLogging)
	r.Use(middleware.RequestLogging)
	r.Use(middleware.SetContentTypeHeader)

	client := initializeRedis(ctx)

	sender := twilio.NewSender(viper.GetString(
		config.TwilioAccountSID),
		viper.GetString(config.TwilioAuthToken),
		viper.GetString(config.TwilioURL),
		viper.GetString(config.TwilioFrom))

	vault, err := vault.New(
		viper.GetString(config.VaultToken),
		viper.GetString(config.VaultUnSealKey),
		viper.GetString(config.VaultAddress),
		viper.GetString(config.UserPath),
		viper.GetString(config.TempPath))
	if err != nil {
		logger.Fatalf(ctx, "router: Error creating vault client: %+v", err)
	}

	fromAccount := &algorand.Account{
		AccountAddress:     viper.GetString(config.FromAddress),
		SecurityPassphrase: viper.GetString(config.FromSecurityParaphrase),
	}

	algo := algorand.New(
		fromAccount,
		viper.GetString(config.ApiAddress),
		viper.GetString(config.ApiKey),
		viper.GetUint64(config.AmountFactor),
		viper.GetUint64(config.MinFee),
		viper.GetUint64(config.SeedAlgo),
	)

	userService := user.NewUser(algo, *vault)
	eventService := event.NewEvent(algo, *vault)
	f := factory.NewFactory()

	r.HandleFunc("/healthcheck", healthcheck.Self).Methods(http.MethodGet)
	baseRouter := r.PathPrefix("/v1").Subrouter()

	userRouter := baseRouter.PathPrefix("/user").Subrouter()
	userRouter.HandleFunc("/connect", handler.CreateUser(userService, f)).Methods(http.MethodPost)
	userRouter.HandleFunc("/connect/verify", handler.VerifyUser(userService, f)).Methods(http.MethodPost)

	marketPlaceRouter := baseRouter.PathPrefix("/marketplace/user").Subrouter()
	marketPlaceRouter.HandleFunc("/connect", handler.CreateMarketPlaceUser(userService, f, sender, client, viper.GetString(config.Secret))).Methods(http.MethodPost)
	marketPlaceRouter.HandleFunc("/verifyotp", handler.VerifyMarketPlaceOTP(userService, f, client)).Methods(http.MethodPost)

	publicEventRouter := baseRouter.PathPrefix("/public_event").Subrouter()
	publicEventRouter.HandleFunc("", handler.PublicEvent(eventService, f)).Methods(http.MethodPost)
	publicEventRouter.HandleFunc("", handler.UpdatePublicEvent(eventService, f)).Methods(http.MethodPatch)
	publicEventRouter.HandleFunc("", handler.GetPublicEvents(eventService, f)).Methods(http.MethodGet)
	publicEventRouter.HandleFunc("/{userID}", handler.GetPublicEvent(eventService, f)).Methods(http.MethodGet)

	return r
}

func initializeRedis(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString(config.RedisAddress),
		Password: viper.GetString(config.RedisPassword),
		DB:       viper.GetInt(config.RedisDB),
	})
	_, err := client.Ping().Result()
	if err != nil {
		logger.Fatalf(ctx, "initializeRedis: error connecting to the Redis DB: %s", err)
	}

	return client
}
