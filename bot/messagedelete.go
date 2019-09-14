package bot

import (
	"github.com/bwmarrin/discordgo"
)

func (b *Bot) messageDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	star := Star{}
	err := b.db.Get(&star, "SELECT * FROM STARS WHERE id = $1", m.ID)
	if err != nil {
		return
	}

	s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)

	b.db.Exec("DELETE FROM stars WHERE id = $1", m.ID)
}
