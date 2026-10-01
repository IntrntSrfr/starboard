package starboard

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

const helpDescription = "There are two settings you can change\n - Starboard channel\n - Minimum required stars for a post to be posted to starboard\n\t - The minimum amount you can set is 1\nTwo examples: \n`/settings edit channel` - Edit the Starboard channel\n`/settings edit minstars 3` - Edit the minimum amount of reactions to appear on Starboard"

type module struct {
	*bot.ModuleBase
	db DB
}

func NewModule(b *bot.Bot, db DB, logger mio.Logger) *module {
	logger = logger.Named("Module")

	return &module{
		ModuleBase: bot.NewModule(b, "commands", logger),
		db:         db,
	}
}

func (m *module) Hook() error {
	if err := m.RegisterApplicationCommands(
		newHelpSlash(m),
		newSettingsSlash(m),
	); err != nil {
		return err
	}

	return nil
}

func newHelpSlash(m *module) *bot.ModuleApplicationCommand {
	cmd := bot.NewModuleApplicationCommandBuilder(m, "help").
		Type(discordgo.ChatApplicationCommand).
		Description("Get help on how to use the bot")

	run := func(d *discord.DiscordApplicationCommand) {
		embed := builders.NewEmbedBuilder().
			WithTitle("Help").
			WithOkColor().
			WithDescription(helpDescription)
		d.RespondEmbed(embed.Build())
	}

	return cmd.Execute(run).Build()
}

func newSettingsSlash(m *module) *bot.ModuleApplicationCommand {
	minStars := minStarsFloor
	handler := NewSettingsCommand(m.db, m.Logger.Named("settings_command"))

	cmd := bot.NewModuleApplicationCommandBuilder(m, "settings").
		Type(discordgo.ChatApplicationCommand).
		Description("View or set settings").
		NoDM().
		Permissions(discordgo.PermissionAdministrator).
		AddSubcommand(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "view",
			Description: "View the current settings",
		}).
		AddSubcommand(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "set",
			Description: "Set a setting",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "stars",
					Description: "Minimum amount of required stars for a message to be posted to the Starboard",
					Required:    true,
					MinValue:    &minStars,
				},
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel for the starboard",
					Required:    true,
				},
			},
		})

	run := func(d *discord.DiscordApplicationCommand) {
		handler.Handle(newDiscordSettingsContext(d))
	}

	return cmd.Execute(run).Build()
}
