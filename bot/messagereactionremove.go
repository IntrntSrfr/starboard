package bot

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (b *Bot) messageReactionRemoveHandler(s *discordgo.Session, r *discordgo.MessageReactionRemove) {

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
		return

	case nil:

		if count < gs.MinStars {
			s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
			b.db.Exec("DELETE FROM stars WHERE id = $1", r.MessageID)
			return
		}

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
