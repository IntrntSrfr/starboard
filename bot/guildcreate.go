package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (b *Bot) guildCreateHandler(s *discordgo.Session, g *discordgo.GuildCreate) {
	sbchannel := ""
	for _, c := range g.Channels {
		if strings.ToLower(c.Name) == "starboard" {
			sbchannel = c.ID
		}
	}

	var count int
	b.db.Get(&count, "SELECT COUNT(*) FROM guildsettings WHERE id = $1;", g.ID)

	if count == 0 {
		if sbchannel != "" {
			s.ChannelMessageSend(sbchannel, "to start using starebord, please check out the help command 'sb.help' for info on how to set it all up")
		}

		_, err := b.db.Exec("INSERT INTO guildsettings VALUES($1, $2, $3);", g.ID, sbchannel, 3)
		if err != nil {
			fmt.Println(err)
			b.logger.Error("error", zap.Error(err))
			return
		}
	}

	owner := ""
	own, err := s.State.Member(g.ID, g.OwnerID)
	if err != nil {
		owner = g.OwnerID
	} else {
		owner = own.User.String()
	}

	b.logger.Info(fmt.Sprintf("LOADED %v - %v", g.Name, owner))
	fmt.Println(fmt.Sprintf("LOADED %v - %v", g.Name, owner))
}
