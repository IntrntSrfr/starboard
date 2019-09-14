package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/intrntsrfr/starboard/bot"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {

	// Create zap logger
	z := zap.NewDevelopmentConfig()
	z.OutputPaths = []string{"./logs.txt"}
	z.ErrorOutputPaths = []string{"./logs.txt"}
	logger, err := z.Build()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer logger.Sync()

	logger.Info("logger construction succeeded")

	// Look for the config file
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
	}

	// Unmashal config
	var config bot.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("error unmarshaling config file")
	}

	// Create database connection
	psql, err := sqlx.Connect("postgres", config.ConnectionString)
	if err != nil {
		panic("could not connect to db")
	}
	logger.Info("Established database connection")

	// Create new discord log bot
	client, err := bot.NewBot(&config, logger.Named("discord"), psql)
	if err != nil {
		return
	}
	defer client.Close()

	// Run the client
	err = client.Run()
	if err != nil {
		return
	}

	// Block forever until ctrl-c
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
