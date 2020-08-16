package bot

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

type Bot struct {
	logger    *zap.Logger
	db        *sqlx.DB
	client    *discordgo.Session
	config    *Config
	starttime time.Time
}

var (
	statusTimer *time.Ticker
)

func NewBot(Config *Config, Log *zap.Logger, psql *sqlx.DB) (*Bot, error) {

	client, err := discordgo.New("Bot " + Config.Token)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	Log.Info("created discord client")

	schemas := []string{
		schemaGuildSettings,
		schemaStars,
	}

	for _, s := range schemas {
		psql.MustExec(s)
	}

	return &Bot{
		logger:    Log,
		db:        psql,
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
	b.client.AddHandler(b.guildCreateHandler)
	b.client.AddHandler(b.messageCreateHandler)
	b.client.AddHandler(b.messageUpdateHandler)
	b.client.AddHandler(b.messageDeleteHandler)
	b.client.AddHandler(b.messageReactionAddHandler)
	b.client.AddHandler(b.messageReactionRemoveHandler)
	b.client.AddHandler(b.messageReactionRemoveAllHandler)
	b.client.AddHandler(b.readyHandler)
	b.client.AddHandler(b.disconnectHandler)
}

func (b *Bot) readyHandler(s *discordgo.Session, r *discordgo.Ready) {

	statusTimer = time.NewTicker(time.Second * 15)

	go func() {
		i := 0
		for range statusTimer.C {
			switch i {
			case 0:
				s.UpdateStatus(0, "sb.help")
				i++
			case 1:
				channels := 0

				for _, g := range s.State.Guilds {
					for _, c := range g.Channels {
						if c.Type == discordgo.ChannelTypeGuildText {
							channels++
						}
					}
				}

				s.UpdateStatusComplex(discordgo.UpdateStatusData{
					Game: &discordgo.Game{
						Name: fmt.Sprintf("%v channels for stars", channels),
						Type: discordgo.GameTypeWatching,
					},
				})
				i++
			default:
				i = 0
			}

		}
	}()

	fmt.Println(fmt.Sprintf("Logged in as %v.", r.User.String()))
}

func (b *Bot) disconnectHandler(s *discordgo.Session, d *discordgo.Disconnect) {
	statusTimer.Stop()
	fmt.Println("DISCONNECTED AT ", time.Now().Format(time.RFC1123))
}
