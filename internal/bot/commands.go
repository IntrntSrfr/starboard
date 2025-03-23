package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"github.com/intrntsrfr/starboard/internal/database"
)

type module struct {
	*bot.ModuleBase
	db database.DB
}

func NewModule(b *bot.Bot, db database.DB, logger mio.Logger) *module {
	logger = logger.Named("Module")

	return &module{
		ModuleBase: bot.NewModule(b, "commands", logger),
		db:         db,
	}
}

func (m *module) Hook() error {
	if err := m.RegisterApplicationCommands(
		newHelpSlash(m),
		//newSettingsSlash(m),
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

		text := strings.Builder{}
		text.WriteString("There are two settings you can change")
		text.WriteString("\n - Starboard channel")
		text.WriteString("\n - Minimum required stars for a post to be posted to starboard")
		text.WriteString("\n\t - The minimum amount you can set is 1")
		text.WriteString("\nTwo examples: ")
		text.WriteString("\n`/settings edit channel` - Edit the Starboard channel")
		text.WriteString("\n`/settings edit minstars 3` - Edit the minimum amount of reactions to appear on Starboard")

		embed := builders.NewEmbedBuilder().
			WithTitle("Help").
			WithOkColor().
			WithDescription(text.String())
		d.RespondEmbed(embed.Build())
	}

	return cmd.Execute(run).Build()
}
