package bot

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func messageUpdateHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageUpdate) {
	return func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		if m.Author == nil || m.Author.Bot || m.GuildID == "" {
			return
		}
		star, err := b.db.GetStar(m.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", zap.Error(err), zap.Any("message", m))
			}
			return
		}
		if starboardMsg, err := s.ChannelMessage(star.StarboardChannelID, star.StarboardMsgID); err == nil {
			if len(starboardMsg.Embeds) < 1 {
				return
			}
			embed := starboardMsg.Embeds[0]
			embed.Description = fmt.Sprintf("%v\n\n\n --> [Jump to message](https://discordapp.com/channels/%v/%v/%v)", m.Content, m.GuildID, m.ChannelID, m.ID)
			_, _ = s.ChannelMessageEditEmbed(star.StarboardChannelID, star.StarboardMsgID, embed)
		}
	}

}
