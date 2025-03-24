package starboard

type Config struct {
	Token            string `json:"token"`
	ConnectionString string `json:"connection_string"`
}

type GuildSettings struct {
	ID                 string `json:"id" db:"id"`
	StarboardChannelID string `json:"starboard_channel_id" db:"starboard_channel_id"`
	MinStars           int    `json:"min_stars" db:"min_stars"`
}

type Star struct {
	ID                 string `json:"id" db:"id"`
	OriginChannelID    string `json:"origin_channel_id" db:"origin_channel_id"`
	StarboardMsgID     string `json:"starboard_msg_id" db:"starboard_msg_id"`
	StarboardChannelID string `json:"starboard_channel_id" db:"starboard_channel_id"`
}

const SchemaGuildSettings = `
CREATE TABLE IF NOT EXISTS guildsettings (
	id                     TEXT PRIMARY KEY,
	starboard_channel_id   TEXT NOT NULL DEFAULT '',
	min_stars              INT NOT NULL DEFAULT 3
);
`

const SchemaStars = `
CREATE TABLE IF NOT EXISTS stars (
	id                   TEXT PRIMARY KEY,
	origin_channel_id    TEXT NOT NULL,
	starboard_msg_id     TEXT NOT NULL,
	starboard_channel_id TEXT NOT NULL
);
`
