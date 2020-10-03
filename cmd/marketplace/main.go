package main

import (
	"context"
	"eventers-marketplace-backend/config"
	c "eventers-marketplace-backend/context"
	"eventers-marketplace-backend/router"
	"flag"
	"fmt"
	l "log"

	"github.com/codegangsta/negroni"
	"github.com/spf13/viper"
)

var (
	version string
)

const defaultCorrelationID = "00000000.00000000"

var ctx context.Context

func init() {
	ctx = c.SetContextWithValue(context.Background(), c.ContextKeyCorrelationID, defaultCorrelationID)
}

func main() {
	cfgPath := flag.String("CONFIG_PATH", "./config.yaml", "Path to config file")
	flag.Parse()

	var err error
	viper.SetConfigFile(*cfgPath)

	err = viper.ReadInConfig()
	if err != nil {
		l.Fatalln("error reading config")
	}

	muxRouter := router.Router(ctx)

	n := negroni.New()
	n.UseHandler(muxRouter)
	n.Run(fmt.Sprintf("%s", viper.GetString(config.Port)))
}
