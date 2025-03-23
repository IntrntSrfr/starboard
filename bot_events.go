package starboard

import (
	"github.com/intrntsrfr/meido/pkg/mio/bot"
)

func logApplicationCommandRan(m *Bot) func(cmd *bot.ApplicationCommandRan) {
	return func(cmd *bot.ApplicationCommandRan) {
		m.logger.Info("Slash",
			"name", cmd.Interaction.Name(),
			"id", cmd.Interaction.ID(),
			"channelID", cmd.Interaction.ChannelID(),
			"userID", cmd.Interaction.AuthorID(),
		)
	}
}

func logApplicationCommandPanicked(m *Bot) func(cmd *bot.ApplicationCommandPanicked) {
	return func(cmd *bot.ApplicationCommandPanicked) {
		m.logger.Error("Slash panic",
			"interaction", cmd.Interaction,
			"reason", cmd.Reason,
		)
	}
}
