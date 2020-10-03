package factory

import (
	"context"
	"database/sql"
	"eventers-marketplace-backend/config"
	"eventers-marketplace-backend/logger"
	"log"
	"sync"

	firebase "firebase.google.com/go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

var db sync.Once
var fa sync.Once

type Factory interface {
	DB(ctx context.Context) *sql.DB
	FirebaseApp(ctx context.Context) *firebase.App
}

type factory struct {
	db  *sql.DB
	app *firebase.App
}

func NewFactory() Factory {
	return &factory{}
}

func (f *factory) DB(ctx context.Context) *sql.DB {
	var dbError error
	db.Do(func() {
		sqlDB, err := sql.Open("mysql", viper.GetString(config.DBURL))
		if err != nil {
			log.Fatal("Error creating connection pool: ", err.Error())
		}

		f.db = sqlDB
		dbError = err
	})

	if dbError != nil {
		logger.Fatalf(ctx, "Could not establish connection to the DB: %+v", dbError)
	}

	return f.db
}

func (f *factory) FirebaseApp(ctx context.Context) *firebase.App {
	var faError error
	fa.Do(func() {
		opt := option.WithCredentialsFile(viper.GetString(config.FirebaseServiceAccountKeyPath))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logger.Fatalf(ctx, "firebaseApp: error initializing firebase app: %+v", err)
		}

		f.app = app
		faError = err
	})

	if faError != nil {
		logger.Fatalf(ctx, "Could not establish connection to the DB: %+v", faError)
	}

	return f.app
}
