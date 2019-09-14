package bot

type Config struct {
	Token            string `json:"token"`
	ConnectionString string `json:"connection_string"`
}

const (
	dColorRed    = 13107200
	dColorOrange = 15761746
	dColorLBlue  = 6410733
	dColorGreen  = 51200
	dColorWhite  = 16777215
)

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

const schemaGuildSettings = `
CREATE TABLE IF NOT EXISTS guildsettings (
	id                     TEXT PRIMARY KEY,
	starboard_channel_id   TEXT,
	min_stars              INT
);
`

const schemaStars = `
CREATE TABLE IF NOT EXISTS stars (
	id                   TEXT PRIMARY KEY,
	origin_channel_id    TEXT,
	starboard_msg_id     TEXT,
	starboard_channel_id TEXT
);
`
