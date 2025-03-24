package starboard

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

const (
	starEmoji = "⭐"
)

var (
	tenorRegex = regexp.MustCompile(`(^https:\/\/media\.tenor\.com\/.*)(AAAAe\/)(.*)(\.png|\.jpg)`)
)

func fixTenorURL(url string) string {
	return tenorRegex.ReplaceAllString(url, "${1}AAAAC/${3}.gif")
}

func getReactionCount(msg *discordgo.Message, emoji string) int {
	for _, r := range msg.Reactions {
		if r.Emoji.Name == emoji {
			return r.Count
		}
	}
	return 0
}

func extractImageURLsAndInfo(msg *discordgo.Message) ([]string, string) {
	imageUrls := make([]string, 0)
	extraContent := ""

	for _, embed := range msg.Embeds {
		if embed.Thumbnail != nil {
			imageUrls = append(imageUrls, fixTenorURL(embed.Thumbnail.URL))
		} else if embed.Image != nil {
			imageUrls = append(imageUrls, fixTenorURL(embed.Image.URL))
		}
	}

	for _, att := range msg.Attachments {
		if att.ContentType == "image/jpeg" || att.ContentType == "image/png" || att.ContentType == "image/gif" {
			imageUrls = append(imageUrls, att.URL)
		}
		extraContent += fmt.Sprintf("\n📎 [%v](%v)", att.Filename, att.URL)
	}

	return imageUrls, extraContent
}

func buildStarboardEmbeds(msg *discordgo.Message, count int, guildID, channelID string) []*discordgo.MessageEmbed {
	content := msg.Content
	jumpLink := fmt.Sprintf("\n\n --> [Jump to message](https://discord.com/channels/%v/%v/%v)\n", guildID, channelID, msg.ID)
	content += jumpLink

	imageUrls, extraContent := extractImageURLsAndInfo(msg)
	content += extraContent

	if len(content) > 4096 {
		content = content[:4096]
	}

	embedBuilder := builders.NewEmbedBuilder().
		WithDescription(content).
		WithOkColor()
	embedBuilder.Author = &discordgo.MessageEmbedAuthor{
		Name:    fmt.Sprintf("%v - ⭐ %v", msg.Author.String(), count),
		IconURL: msg.Author.AvatarURL("64"),
	}
	embeds := []*discordgo.MessageEmbed{embedBuilder.Build()}

	if len(imageUrls) > 0 {
		embedBuilder.WithImageUrl(imageUrls[0])
		embedBuilder.WithUrl(imageUrls[0])

		for _, url := range imageUrls[1:] {
			additionalEmbed := builders.NewEmbedBuilder().
				WithImageUrl(url).
				WithUrl(imageUrls[0]).
				Build()
			embeds = append(embeds, additionalEmbed)
		}
	}

	return embeds
}

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
		if r.Emoji.Name != starEmoji {
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
		count := getReactionCount(msg, starEmoji)
		star, err := b.db.GetStar(r.MessageID)
		switch err {
		case sql.ErrNoRows:
			if count < gs.MinStars {
				return
			}

			embeds := buildStarboardEmbeds(msg, count, r.GuildID, r.ChannelID)
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
		if r.Emoji.Name != starEmoji {
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
		count := getReactionCount(msg, starEmoji)
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
