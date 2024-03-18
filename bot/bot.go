package main

import (
	"log"
	"time"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"gorm.io/gorm"
)

var (
	club48ChatId int64 = -1001345282090
	bot          *gotgbot.Bot
	adminUserId  int64 = 1929567913
	mySql        *gorm.DB
)

func main() {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)
	dispatcher.AddHandler(handlers.NewCommand("start", commandStart))
	dispatcher.AddHandler(handlers.NewCommand("export", commandExport))
	dispatcher.AddHandler(handlers.NewCommand("sum", commandSum))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, messageEcho))
	err := updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", bot.User.Username)
	// Idle, to keep updates coming in, and avoid bot stopping.

	sync()

	updater.Idle()
}
