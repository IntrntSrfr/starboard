package database

import "github.com/intrntsrfr/starboard/internal/structs"

type DB interface {
	Init() error
	Close() error

	GuildDB
	StarDB
}

type GuildDB interface {
	CreateGuild(guildID string) error
	GetGuild(guildID string) (*structs.GuildSettings, error)
	UpdateGuild(g *structs.GuildSettings) error
}

type StarDB interface {
	CreateStar(messageID, channelID, botMessageID, starboardChannelID string) error
	GetStar(messageID string) (*structs.Star, error)
	UpdateStar(star *structs.Star) error
	DeleteStar(messageID string) error
}
