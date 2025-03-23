package starboard

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"
)

func messageDeleteHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageDelete) {
	return func(s *discordgo.Session, m *discordgo.MessageDelete) {
		star, err := b.db.GetStar(m.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(m.ID); err != nil {
			b.logger.Error("could not delete star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
			return
		}
	}
}

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
				att := msg.Attachments[0]
				switch att.ContentType {
				case "image/jpeg", "image/png", "image/gif":
					embed.WithImageUrl(att.URL)
				default:
					embed.Description = fmt.Sprintf("[[%v](%v)]", att.Filename, att.URL)
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

func messageReactionRemoveAllHandler(b *Bot) func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
		star, err := b.db.GetStar(r.MessageID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", zap.Error(err), zap.Any("react remove all", r))
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(r.MessageID); err != nil {
			b.logger.Error("could not delete star", zap.Error(err), zap.Any("react remove all", r))
		}
	}
}

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

func guildCreateHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := b.db.GetGuild(g.ID); err != nil {
			b.logger.Debug("get guild", zap.Error(err))
			if err != sql.ErrNoRows {
				b.logger.Error("could not get guild", zap.Error(err), zap.Any("guild", g))
				return
			}
			if err = b.db.CreateGuild(g.ID); err != nil {
				b.logger.Error("could not create guild", zap.Error(err), zap.Any("guild", g))
				return
			}
		}
		b.logger.Info("guild create", zap.String("id", g.ID))
	}
}
