package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/starboard"
	_ "github.com/lib/pq"
)

func main() {
	cfg := utils.NewConfig()
	loadConfig(cfg, "./config.json")

	psql, err := starboard.NewPSQLDatabase(cfg.GetString("connection_string"))
	if err != nil {
		panic(err)
	}

	bot := starboard.NewBot(cfg, psql)
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

	var c starboard.Config
	if err := json.Unmarshal(f, &c); err != nil {
		panic(err)
	}

	cfg.Set("token", c.Token)
	cfg.Set("shards", 1)
	cfg.Set("connection_string", c.ConnectionString)
}
