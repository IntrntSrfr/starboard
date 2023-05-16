package bot

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/starboard/database"
)

type Bot struct {
	logger    *zap.Logger
	db        database.DB
	client    *discordgo.Session
	config    *Config
	starttime time.Time
}

func NewBot(Config *Config, Log *zap.Logger, db database.DB) (*Bot, error) {
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

func (b *Bot) Close() {
	b.logger.Info("Shutting down bot")
	b.db.Close()
	b.client.Close()
}

func (b *Bot) Run() error {
	b.addHandlers()
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	return b.client.Open()
}

func (b *Bot) addHandlers() {
	b.client.AddHandlerOnce(statusLoop(b))
	b.client.AddHandler(disconnectHandler(b))
	b.client.AddHandler(b.guildCreateHandler)
	b.client.AddHandler(b.messageCreateHandler)
	b.client.AddHandler(b.messageUpdateHandler)
	b.client.AddHandler(b.messageDeleteHandler)
	b.client.AddHandler(b.messageReactionAddHandler)
	b.client.AddHandler(b.messageReactionRemoveHandler)
	b.client.AddHandler(b.messageReactionRemoveAllHandler)
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
					name = "sb.help"
					statusType = discordgo.ActivityTypeGame
				case 1:
					name = fmt.Sprintf("for %v", time.Since(b.starttime).String())
					statusType = discordgo.ActivityTypeGame
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
