package database

import "github.com/intrntsrfr/starboard/bot"

type DB interface {
	Init() error
	Close() error

	GuildDB
	StarDB
}

type GuildDB interface {
	CreateGuild(guildID string) error
	GetGuild(guildID string) (*bot.GuildSettings, error)
	UpdateGuild(g *bot.GuildSettings) error
}

type StarDB interface {
	CreateStar(messageID string) error
	GetStar(messageID string) (*bot.Star, error)
	UpdateStar(star *bot.Star) error
	DeleteStar(messageID string) error
}
