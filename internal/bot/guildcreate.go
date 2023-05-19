package bot

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func guildCreateHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := b.db.GetGuild(g.ID); err != nil {
			b.logger.Debug("get guild", zap.Error(err))
			if err != sql.ErrNoRows {
				b.logger.Error("could not get guild", zap.Error(err), zap.Any("guild", g))
				return
			}
			if err = b.db.CreateGuild(g.ID); err != nil {
				b.logger.Error("could not create guild", zap.Error(err), zap.Any("guild", g))
				return
			}
		}
		b.logger.Info("guild create", zap.String("id", g.ID))
	}
}
