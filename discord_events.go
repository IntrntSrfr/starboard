package starboard

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

var (
	tenorRegex = regexp.MustCompile(`(^https:\/\/media\.tenor\.com\/.*)(AAAAe\/)(.*)(\.png|\.jpg)`)
)

func messageDeleteHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageDelete) {
	return func(s *discordgo.Session, m *discordgo.MessageDelete) {
		star, err := b.db.GetStar(m.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", "error", err, "event", "message delete", "message", m)
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(m.ID); err != nil {
			b.logger.Error("could not delete star", "error", err, "event", "message delete", "message", m)
			return
		}
	}
}

func messageReactionAddHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if u, err := b.Bot.Discord.Member(r.GuildID, r.UserID); err != nil || u.User.Bot {
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

			content := msg.Content
			extraContent := ""
			extraContent += fmt.Sprintf("\n\n --> [Jump to message](https://discordapp.com/channels/%v/%v/%v)", r.GuildID, r.ChannelID, r.MessageID)

			imageUrls := make([]string, 0)
			if len(msg.Embeds) > 0 {
				// fetch embeds with thumbnail or image
				for _, e := range msg.Embeds {
					if e.Thumbnail == nil && e.Image == nil {
						continue
					}
					if e.Thumbnail != nil {
						imageUrls = append(imageUrls, e.Thumbnail.URL)
						continue
					}
					imageUrls = append(imageUrls, e.Image.URL)
				}

				// regex fix any tenor gif links
				for i, url := range imageUrls {
					fmt.Println("url", url)
					result := tenorRegex.ReplaceAllString(url, "${1}AAAAC/${3}.gif")
					imageUrls[i] = result
					fmt.Println("result", result)
				}
			}

			for _, att := range msg.Attachments {
				if att.ContentType == "image/jpeg" || att.ContentType == "image/png" || att.ContentType == "image/gif" {
					imageUrls = append(imageUrls, att.URL)
				}
				extraContent += fmt.Sprintf("\n📎 [%v](%v)", att.Filename, att.URL)
			}

			content += extraContent
			if len(content) > 4096 {
				content = content[:4096]
			}

			embed := builders.NewEmbedBuilder().
				WithDescription(content).
				WithOkColor()

			embed.Author = &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("%v - ⭐ %v", msg.Author.String(), count),
				IconURL: msg.Author.AvatarURL("64"),
			}

			embeds := []*discordgo.MessageEmbed{embed.Build()}

			// add first image to first embed, rest to new embeds.
			// they get combined into one embed
			doneFirst := false
			for _, url := range imageUrls {
				if !doneFirst {
					doneFirst = true
					embed.WithImageUrl(url)
					embed.WithUrl(url)
					continue
				}

				embed := builders.NewEmbedBuilder().
					WithImageUrl(url).
					WithUrl(url).
					Build()
				embeds = append(embeds, embed)
			}

			sentMsg, err := s.ChannelMessageSendComplex(gs.StarboardChannelID, &discordgo.MessageSend{Embeds: embeds})
			if err != nil {
				return
			}

			if err := b.db.CreateStar(r.MessageID, r.ChannelID, sentMsg.ID, gs.StarboardChannelID); err != nil {
				b.logger.Error("could not create star", "error", err)
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
			b.logger.Error("error", "error", err)
		}
	}
}

func messageReactionRemoveHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if u, err := b.Bot.Discord.Member(r.GuildID, r.UserID); err != nil || u.User.Bot {
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
					b.logger.Error("could not delete starboard message", "error", err)
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
			b.logger.Error("error", "error", err)
		}
	}
}

func messageReactionRemoveAllHandler(b *Bot) func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
	return func(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {
		star, err := b.db.GetStar(r.MessageID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", "error", err, "react remove all", r)
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(r.MessageID); err != nil {
			b.logger.Error("could not delete star", "error", err, "react remove all", r)
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
				b.logger.Error("could not get star", "error", err, "message", m)
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
			b.logger.Debug("get guild", "error", err)
			if err != sql.ErrNoRows {
				b.logger.Error("could not get guild", "error", err, "guild", g)
				return
			}
			if err = b.db.CreateGuild(g.ID); err != nil {
				b.logger.Error("could not create guild", "error", err, "guild", g)
				return
			}
		}
		b.logger.Info("guild create", "id", g.ID)
	}
}
