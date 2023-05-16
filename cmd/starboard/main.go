package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/intrntsrfr/starboard/bot"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerConfig.Build()
	logger = logger.Named("main")

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
