package database

import (
	"fmt"
	"github.com/intrntsrfr/starboard/internal/structs"
	"github.com/jmoiron/sqlx"
)

type PsqlDB struct {
	pool *sqlx.DB
}

func NewPSQLDatabase(connStr string) (*PsqlDB, error) {
	pool, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db := &PsqlDB{pool}
	err = db.Init()
	return db, err
}

func (db *PsqlDB) Init() error {
	schemas := map[string]string{
		"guild": structs.SchemaGuildSettings,
		"stars": structs.SchemaStars,
	}

	for k, v := range schemas {
		fmt.Println("executing schema:", k)
		if _, err := db.pool.Exec(v); err != nil {
			return err
		}
	}
	return nil
}

func (db *PsqlDB) Close() error {
	return db.pool.Close()
}

func (db *PsqlDB) CreateStar(messageID, channelID, botMessageID, starboardChannelID string) error {
	_, err := db.pool.Exec("INSERT INTO stars VALUES ($1, $2, $3, $4);",
		messageID, channelID, botMessageID, starboardChannelID)
	return err
}

func (db *PsqlDB) GetStar(messageID string) (*structs.Star, error) {
	var star structs.Star
	err := db.pool.Get(&star, "SELECT * FROM stars WHERE id = $1;", messageID)
	return &star, err
}

func (db *PsqlDB) UpdateStar(star *structs.Star) error {
	_, err := db.pool.Exec("UPDATE stars SET origin_channel_id=$1, starboard_msg_id=$2, starboard_channel_id=$3 WHERE id=$4",
		star.OriginChannelID, star.StarboardMsgID, star.StarboardChannelID, star.ID)
	return err
}

func (db *PsqlDB) DeleteStar(messageID string) error {
	_, err := db.pool.Exec("DELETE FROM stars WHERE id = $1;", messageID)
	return err
}

func (db *PsqlDB) CreateGuild(guildID string) error {
	_, err := db.pool.Exec("INSERT INTO guildsettings VALUES($1, $2, $3);", guildID, "", 3)
	return err
}

func (db *PsqlDB) GetGuild(guildID string) (*structs.GuildSettings, error) {
	var guild structs.GuildSettings
	err := db.pool.Get(&guild, "SELECT * FROM guildsettings WHERE id = $1;", guildID)
	return &guild, err
}

func (db *PsqlDB) UpdateGuild(g *structs.GuildSettings) error {
	_, err := db.pool.Exec("UPDATE guildsettings SET min_stars=$1, starboard_channel_id=$2 WHERE id=$3",
		g.MinStars, g.StarboardChannelID, g.ID)
	return err
}
