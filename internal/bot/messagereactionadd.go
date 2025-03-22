package bot

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"
)

func messageReactionAddHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
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
			if count < gs.MinStars {
				return
			}
			embed := builders.NewEmbedBuilder().
				WithAuthor(fmt.Sprintf("%v - ⭐ %v", msg.Author.String(), count), msg.Author.AvatarURL("64")).
				WithDescription(msg.Content).
				WithOkColor().
				WithTimestamp(time.Now().Format(time.RFC3339)).
				AddField("Author", msg.Author.Mention(), true).
				AddField("Channel", fmt.Sprintf("<#%v>", r.ChannelID), true)

			if len(msg.Attachments) > 0 {
				ext := filepath.Ext(msg.Attachments[0].Filename)
				switch ext {
				case ".gif", ".png", ".jpg":
					embed.WithImageUrl(msg.Attachments[0].URL)
				default:
					embed.Description = fmt.Sprintf("[[%v](%v)]", msg.Attachments[0].Filename, msg.Attachments[0].URL)
				}
			}
			embed.Description += fmt.Sprintf("\n\n\n --> [Jump to message](https://discordapp.com/channels/%v/%v/%v)", r.GuildID, r.ChannelID, r.MessageID)
			starboardMsg, err := s.ChannelMessageSendEmbed(gs.StarboardChannelID, embed.Build())
			if err != nil {
				return
			}
			if err := b.db.CreateStar(r.MessageID, r.ChannelID, starboardMsg.ID, gs.StarboardChannelID); err != nil {
				b.logger.Error("could not create star", zap.Error(err))
				return
			}
		case nil:
			starboardMsg, err := s.ChannelMessage(star.StarboardChannelID, star.StarboardMsgID)
			if err != nil || len(starboardMsg.Embeds) < 1 {
				return
			}

			embed := starboardMsg.Embeds[0]
			embed.Author.Name = fmt.Sprintf("%v - ⭐ %v", msg.Author.String(), count)
			_, _ = s.ChannelMessageEditEmbed(star.StarboardChannelID, star.StarboardMsgID, embed)
		default:
			b.logger.Error("error", zap.Error(err))
		}
	}
}
