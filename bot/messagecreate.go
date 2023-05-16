package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func (b *Bot) messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	g, err := s.State.Guild(m.GuildID)
	if err != nil {
		b.logger.Error("could not fetch guild", zap.Error(err), zap.String("guild id", m.GuildID))
		return
	}

	ch, err := s.State.Channel(m.ChannelID)
	if err != nil {
		b.logger.Info("could not fetch channel", zap.Error(err), zap.String("channel id", m.ChannelID))
		return
	}

	if ch.Type != discordgo.ChannelTypeGuildText {
		return
	}

	args := strings.Fields(strings.ToLower(m.Content))
	uperms, err := s.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		return
	}

	if args[0] == "sb.set" {
		if len(args) < 2 {
			return
		}

		if uperms&discordgo.PermissionManageMessages == 0 || uperms&discordgo.PermissionAdministrator == 0 {
			s.ChannelMessageSend(m.ChannelID, "you need manage message / admin perms to do this")
			return
		}

		switch args[1] {
		case "starboard":
			b.db.Exec("UPDATE guildsettings SET starboard_channel_id = $1 WHERE id = $2", ch.ID, g.ID)
			s.ChannelMessageSend(ch.ID, fmt.Sprintf("Starboard channel is now set to <#%v>", ch.ID))
		case "minstars":
			if len(args) < 3 {
				return
			}
			min, err := strconv.Atoi(args[2])
			if err != nil {
				return
			}
			if min < 1 {
				min = 1
			}
			b.db.Exec("UPDATE guildsettings SET min_stars = $1 WHERE id = $2", min, g.ID)
			s.ChannelMessageSend(ch.ID, fmt.Sprintf("Required stars is now set to %v", min))

		default:
		}

	} else if args[0] == "sb.help" {
		text := strings.Builder{}
		text.WriteString("There are two settings you can change")
		text.WriteString("\n - Starboard channel")
		text.WriteString("\n - Minimum required stars for a post to be posted to starboard")
		text.WriteString("\n\t - The minimum amount you can set is 1")
		text.WriteString("\nTwo examples: ")
		text.WriteString("\n`sb.set starboard` - this will set starboard to whatever channel the command is posted")
		text.WriteString("\n`sb.set minstars 3` - this will set the required stars to 3")
		s.ChannelMessageSend(ch.ID, text.String())
	}
}
