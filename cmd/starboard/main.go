package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/starboard/internal/bot"
	"github.com/intrntsrfr/starboard/internal/database"
	"github.com/intrntsrfr/starboard/internal/structs"
	_ "github.com/lib/pq"
)

func main() {
	cfg := utils.NewConfig()
	loadConfig(cfg, "./config.json")

	psql, err := database.NewPSQLDatabase(cfg.GetString("connection_string"))
	if err != nil {
		panic(err)
	}

	bot := bot.NewBot(cfg, psql)
	defer bot.Close()

	if err = bot.Run(context.Background()); err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}

func loadConfig(cfg *utils.Config, path string) {
	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var c structs.Config
	if err := json.Unmarshal(f, &c); err != nil {
		panic(err)
	}

	cfg.Set("token", c.Token)
	cfg.Set("shards", 1)
	cfg.Set("connection_string", c.ConnectionString)
}
