package starboard

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

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

func newSettingsSlash(m *module) *bot.ModuleApplicationCommand {
	minStars := 1.0

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
		gc, err := m.db.GetGuild(d.GuildID())
		if err != nil {
			d.Respond("Couldn't get server config")
			m.Logger.Error("Couldn't get server config", "error", err)
			return
		}

		if _, ok := d.Options("view"); ok {
			d.RespondEmbed(generateSettingsEmbed(gc))
			return
		} else if _, ok := d.Options("set"); ok {
			// required setting, dont bother checking if it exists
			starsOpt, _ := d.Options("set:stars")
			stars := starsOpt.IntValue()

			// required setting, dont bother checking if it exists
			chOpt, _ := d.Options("set:channel")
			ch := chOpt.ChannelValue(d.Sess.Real())

			if ch == nil {
				d.Respond("Couldn't find that channel")
				return
			}

			gc.MinStars = int(stars)
			gc.StarboardChannelID = ch.ID

			if err := m.db.UpdateGuild(gc); err != nil {
				d.Respond("Couldn't update server config")
				m.Logger.Error("Couldn't get server config", "error", err)
				return
			}

			embed := generateSettingsEmbed(gc)
			embed.Title = "Updated settings"

			resp := &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Flags:  discordgo.MessageFlagsEphemeral,
			}

			d.RespondComplex(resp, discordgo.InteractionResponseChannelMessageWithSource)
			return
		}
	}

	return cmd.Execute(run).Build()
}

func generateSettingsEmbed(gc *GuildSettings) *discordgo.MessageEmbed {
	channelFieldStr := fmt.Sprintf("<#%v>", gc.StarboardChannelID)
	if gc.StarboardChannelID == "" {
		channelFieldStr = "Not set"
	}

	embed := builders.NewEmbedBuilder().
		WithTitle("Settings").
		WithOkColor().
		AddField("Stars required", fmt.Sprint(gc.MinStars), true).
		AddField("Starboard channel", channelFieldStr, true)

	return embed.Build()
}
