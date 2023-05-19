package bot

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func messageDeleteHandler(b *Bot) func(s *discordgo.Session, m *discordgo.MessageDelete) {
	return func(s *discordgo.Session, m *discordgo.MessageDelete) {
		star, err := b.db.GetStar(m.ID)
		if err != nil {
			if err != sql.ErrNoRows {
				b.logger.Error("could not get star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
			}
			return
		}
		_ = s.ChannelMessageDelete(star.StarboardChannelID, star.StarboardMsgID)
		if err = b.db.DeleteStar(m.ID); err != nil {
			b.logger.Error("could not delete star", zap.Error(err), zap.String("event", "message delete"), zap.Any("message", m))
			return
		}
	}
}
