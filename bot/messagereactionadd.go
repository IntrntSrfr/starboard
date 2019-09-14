package bot

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) messageReactionAddHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {

	u, err := s.State.Member(r.GuildID, r.UserID)
	if err != nil {
		return
	}

	if u.User.Bot {
		return
	}

	if r.Emoji.Name != "⭐" {
		return
	}

	gs := GuildSettings{}
	err = b.db.Get(&gs, "SELECT * FROM guildsettings WHERE id = $1", r.GuildID)
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

	star := Star{}
	err = b.db.Get(&star, "SELECT * FROM stars WHERE id = $1", r.MessageID)

	switch err {
	case sql.ErrNoRows:
		if count >= gs.MinStars {

			// post to starboard
			u, err := s.State.Member(r.GuildID, r.UserID)
			if err != nil {
				return
			}

			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					IconURL: msg.Author.AvatarURL("64"),
					Name:    msg.Author.String() + fmt.Sprintf(" - ⭐ %v", count),
				},
				Description: msg.Content,
				Timestamp:   time.Now().Format(time.RFC3339),
				Color:       dColorWhite,
				//Title: fmt.Sprintf("⭐ %v", count),
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Author",
						Value:  u.Mention(),
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Channel",
						Value:  "<#" + r.ChannelID + ">",
						Inline: true,
					},
				},
			}

			if len(msg.Attachments) > 0 {
				file := strings.Split(msg.Attachments[0].Filename, ".")
				extension := file[len(file)-1]

				switch extension {
				case "gif", "png", "jpg":
					embed.Image = &discordgo.MessageEmbedImage{
						URL: msg.Attachments[0].URL,
					}
				default:
					embed.Description = fmt.Sprintf("[[%v](%v)]", msg.Attachments[0].Filename, msg.Attachments[0].URL)
				}
			}

			embed.Description += fmt.Sprintf("\n\n\n --> [Jump to message](https://discordapp.com/channels/%v/%v/%v)", r.GuildID, r.ChannelID, r.MessageID)

			starboardMsg, err := s.ChannelMessageSendEmbed(gs.StarboardChannelID, embed)
			if err != nil {
				return
			}

			b.db.Exec(
				"INSERT INTO stars VALUES ($1, $2, $3, $4)",
				r.MessageID, r.ChannelID, starboardMsg.ID, gs.StarboardChannelID,
			)
		}
	case nil:
		starboardMsg, err := s.ChannelMessage(star.StarboardChannelID, star.StarboardMsgID)
		if err != nil {
			return
		}
		if len(starboardMsg.Embeds) < 1 {
			return
		}

		embed := starboardMsg.Embeds[0]
		embed.Author.Name = msg.Author.String() + fmt.Sprintf(" - ⭐ %v", count)
		//editEmbed.Title = fmt.Sprintf("⭐ %v", count)
		s.ChannelMessageEditEmbed(star.StarboardChannelID, star.StarboardMsgID, embed)

	default:
		b.logger.Error("error", zap.Error(err))
	}
}
