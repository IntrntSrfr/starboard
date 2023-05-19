package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/intrntsrfr/starboard/internal/bot"
	"github.com/intrntsrfr/starboard/internal/database"
	"github.com/intrntsrfr/starboard/internal/structs"
	_ "github.com/lib/pq"
)

func main() {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerConfig.Build()
	logger = logger.Named("main")

	file, err := os.ReadFile("./config.json")
	if err != nil {
		logger.Panic("could not read file", zap.Error(err))
	}

	var config structs.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		logger.Panic("could not unmarshal config", zap.Error(err))
	}

	psql, err := database.NewPSQLDatabase(config.ConnectionString)
	if err != nil {
		logger.Panic("could not connect to db", zap.Error(err))
	}

	// Create new discord log bot
	client, err := bot.NewBot(&config, logger.Named("discord"), psql)
	if err != nil {
		return
	}
	defer closeBot(client, logger)

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

func closeBot(client *bot.Bot, log *zap.Logger) {
	err := client.Close()
	if err != nil {
		log.Error("could not close bot", zap.Error(err))
	}
}
