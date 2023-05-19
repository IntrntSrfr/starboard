package bot

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/starboard/internal/database"
	"github.com/intrntsrfr/starboard/internal/structs"
)

type Bot struct {
	logger    *zap.Logger
	db        database.DB
	client    *discordgo.Session
	config    *structs.Config
	starttime time.Time
}

func NewBot(Config *structs.Config, Log *zap.Logger, db database.DB) (*Bot, error) {
	client, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	intents := discordgo.IntentsAllWithoutPrivileged |
		discordgo.IntentGuildMembers |
		discordgo.IntentGuildPresences |
		discordgo.IntentMessageContent
	client.Identify.Intents = discordgo.MakeIntent(intents)

	return &Bot{
		logger:    Log,
		db:        db,
		client:    client,
		config:    Config,
		starttime: time.Now(),
	}, nil
}

func (b *Bot) Close() error {
	b.logger.Info("shutdown")
	if err := b.client.Close(); err != nil {
		return err
	}
	if err := b.db.Close(); err != nil {
		return err
	}
	return nil
}

func (b *Bot) Run() error {
	b.addHandlers()
	b.logger.Info("starting")
	return b.client.Open()
}

func (b *Bot) addHandlers() {
	b.client.AddHandlerOnce(statusLoop(b))
	b.client.AddHandler(interactionCreate(b))
	b.client.AddHandler(disconnectHandler(b))
	b.client.AddHandler(guildCreateHandler(b))
	//b.client.AddHandler(b.messageCreateHandler)
	b.client.AddHandler(messageUpdateHandler(b))
	b.client.AddHandler(messageDeleteHandler(b))
	b.client.AddHandler(messageReactionAddHandler(b))
	b.client.AddHandler(messageReactionRemoveHandler(b))
	b.client.AddHandler(messageReactionRemoveAllHandler(b))
}

const totalStatusDisplays = 2

func statusLoop(b *Bot) func(s *discordgo.Session, r *discordgo.Ready) {
	statusTimer := time.NewTicker(time.Second * 15)
	return func(s *discordgo.Session, r *discordgo.Ready) {
		display := 0
		go func() {
			for range statusTimer.C {
				var (
					name       string
					statusType discordgo.ActivityType
				)
				switch display {
				case 0:
					name = "/help"
					statusType = discordgo.ActivityTypeGame
				case 1:
					name = "you forever O_O"
					statusType = discordgo.ActivityTypeWatching
				}
				_ = s.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: name,
						Type: statusType,
					}},
				})
				display = (display + 1) % totalStatusDisplays
			}
		}()
	}
}

func disconnectHandler(b *Bot) func(s *discordgo.Session, d *discordgo.Disconnect) {
	return func(s *discordgo.Session, d *discordgo.Disconnect) {
		b.logger.Warn("disconnected")
	}
}
