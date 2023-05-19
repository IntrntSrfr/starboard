package bot

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func messageReactionRemoveHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if u, err := s.State.Member(r.GuildID, r.UserID); err != nil || u.User.Bot {
			return
		}
		if r.Emoji.Name != "⭐" {
			return
		}
		gs, err := b.db.GetGuild(r.GuildID)
		if err != nil {
			return
		}
		msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
		if err != nil {
			return
		}
		count := 0
		for _, mr := range msg.Reactions {
			if mr.Emoji.Name == "⭐" {
				count = mr.Count
			}
		}
		star, err := b.db.GetStar(r.MessageID)
		switch err {
		case sql.ErrNoRows:
			return
		case nil:
			if count < gs.MinStars {
				_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
				if err = b.db.DeleteStar(r.MessageID); err != nil {
					b.logger.Error("could not deleted", zap.Error(err))
				}
				return
			}
			starboardMsg, err := s.ChannelMessage(star.StarboardChannelID, star.StarboardMsgID)
			if err != nil || len(starboardMsg.Embeds) < 1 {
				return
			}
			embed := starboardMsg.Embeds[0]
			embed.Author.Name = msg.Author.String() + fmt.Sprintf(" - ⭐ %v", count)
			_, _ = s.ChannelMessageEditEmbed(star.StarboardChannelID, star.StarboardMsgID, embed)
		default:
			b.logger.Error("error", zap.Error(err))
		}
	}
}
