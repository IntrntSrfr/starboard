package bot

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (b *Bot) messageUpdateHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	gs := GuildSettings{}
	err := b.db.Get(&gs, "SELECT * FROM guildsettings WHERE id = $1", m.GuildID)
	if err != nil {
		return
	}

	star := Star{}
	err = b.db.Get(&star, "SELECT * FROM stars WHERE id = $1", m.ID)

	switch err {
	case sql.ErrNoRows:
		return

	case nil:
		starboardMsg, err := s.ChannelMessage(star.StarboardChannelID, star.StarboardMsgID)
		if err != nil {
			return
		}
		if len(starboardMsg.Embeds) < 1 {
			return
		}

		embed := starboardMsg.Embeds[0]
		embed.Description = fmt.Sprintf("%v\n\n\n --> [Jump to message](https://discordapp.com/channels/%v/%v/%v)", m.Content, m.GuildID, m.ChannelID, m.ID)

		s.ChannelMessageEditEmbed(star.StarboardChannelID, star.StarboardMsgID, embed)

		//editEmbed.Title = fmt.Sprintf("⭐ %v", count)
	default:
		b.logger.Error("error", zap.Error(err))

	}
}
