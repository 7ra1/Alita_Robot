package modules

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/divkix/Alita_Robot/alita/utils/helpers"
)

var (
	GambleModule = moduleStruct{moduleName: "Gamble"}
)

// ping handles the /ping command to measure bot-to-Telegram API round-trip time
func (moduleStruct) Gamble(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Step 3: Edit with detailed breakdown
	text := fmt.Sprintf("Gambled! 🎲\n\nRound-trip time: %d ms", msg.Date-time.Now().Unix())
	sentMsg, err := msg.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: helpers.HTML,
	})
	text = fmt.Sprintf("Gambled! 🎲\n\nRound-trip time: %d ms", msg.Date-time.Now().Unix())

	sentMsg.EditText(b, text, &gotgbot.EditMessageTextOpts{
		ParseMode: helpers.HTML,
	})
	if err != nil {
		log.WithError(err).Error("[Ping] Failed to edit ping response")
		return err
	}

	return ext.EndGroups
}

// LoadMisc registers all miscellaneous module handlers with the dispatcher,
// including utility commands for IDs, ping, translation, and stats.
func LoadGamble(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(miscModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("ping", miscModule.ping))
	helpers.AddCmdToDisableable("ping")

}
