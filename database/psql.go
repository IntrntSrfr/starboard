package database

import (
	"github.com/intrntsrfr/starboard/bot"
	"github.com/jmoiron/sqlx"
)

type PsqlDB struct {
	pool *sqlx.DB
}

func (db *PsqlDB) Init() error {
	schemas := []string{
		bot.SchemaGuildSettings,
		bot.SchemaStars,
	}

	for _, s := range schemas {
		_, err := db.pool.Exec(s)
		return err
	}
	return nil
}

func (db *PsqlDB) Close() error {
	return db.pool.Close()
}

func (db *PsqlDB) CreateStar(messageID, channelID, botMessageID, starboardChannelID string) error {
	_, err := db.pool.Exec("INSERT INTO stars VALUES ($1, $2, $3, $4)",
		messageID, channelID, botMessageID, starboardChannelID)
	return err
}

func (db *PsqlDB) GetStar(messageID string) (*bot.Star, error) {
	var star bot.Star
	err := db.pool.Get(&star, "SELECT * FROM STARS WHERE id = $1", messageID)
	return &star, err
}

func (db *PsqlDB) UpdateStar(star *bot.Star) error {
	panic("not implemented") // TODO: Implement
}

func (db *PsqlDB) DeleteStar(messageID string) error {
	_, err := db.pool.Exec("DELETE FROM stars WHERE id = $1", messageID)
	return err
}

func (db *PsqlDB) CreateGuild(guildID string) error {
	_, err := db.pool.Exec("INSERT INTO guildsettings VALUES($1, $2, $3);", guildID, "", 3)
	return err
}

func (db *PsqlDB) GetGuild(guildID string) (*bot.GuildSettings, error) {
	var guild bot.GuildSettings
	err := db.pool.Get("SELECT * FROM guildsettings WHERE id = $1", guildID)
	return &guild, err
}

func (db *PsqlDB) UpdateGuild(g *bot.GuildSettings) error {
	panic("not implemented") // TODO: Implement
}
