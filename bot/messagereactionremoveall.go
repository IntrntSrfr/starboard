package bot

import "github.com/bwmarrin/discordgo"

func (b *Bot) messageReactionRemoveAllHandler(s *discordgo.Session, r *discordgo.MessageReactionRemoveAll) {

	star := Star{}
	err := b.db.Get(&star, "SELECT * FROM STARS WHERE id = $1", r.MessageID)
	if err != nil {
		return
	}

	s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)

	b.db.Exec("DELETE FROM stars WHERE id = $1", r.MessageID)
}
