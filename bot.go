package starboard

import (
	"context"

	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/utils"
)

type Bot struct {
	Bot    *bot.Bot
	logger mio.Logger
	db     DB
	config *utils.Config
}

func NewBot(config *utils.Config, db DB) *Bot {
	logger := mio.NewDefaultLogger().Named("Bot")

	b := bot.NewBotBuilder(config).
		WithDefaultHandlers().
		WithLogger(logger).
		Build()

	return &Bot{
		Bot:    b,
		logger: logger,
		config: config,
		db:     db,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	b.registerModules()
	b.registerDiscordHandlers()
	b.registerMioHandlers()
	return b.Bot.Run(ctx)
}

func (b *Bot) Close() {
	b.Bot.Close()
}

func (b *Bot) registerModules() {
	modules := []bot.Module{
		NewModule(b.Bot, b.db, b.logger),
	}
	for _, mod := range modules {
		b.Bot.RegisterModule(mod)
	}
}

func (b *Bot) registerDiscordHandlers() {
	b.Bot.Discord.AddEventHandler(guildCreateHandler(b))
	b.Bot.Discord.AddEventHandler(messageUpdateHandler(b))
	b.Bot.Discord.AddEventHandler(messageDeleteHandler(b))
	b.Bot.Discord.AddEventHandler(messageReactionAddHandler(b))
	b.Bot.Discord.AddEventHandler(messageReactionRemoveHandler(b))
	b.Bot.Discord.AddEventHandler(messageReactionRemoveAllHandler(b))
}

func (b *Bot) registerMioHandlers() {
	b.Bot.AddHandler(logApplicationCommandPanicked(b))
	b.Bot.AddHandler(logApplicationCommandRan(b))
}
