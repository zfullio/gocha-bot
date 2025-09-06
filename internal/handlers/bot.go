package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func RunApp(bh *th.BotHandler) {
	bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		handleWebAppCommand(ctx, message.Chat.ID)

		return nil
	}, th.CommandEqual("start"))
}

func handleWebAppCommand(ctx *th.Context, chatID int64) {
	menu := &telego.InlineKeyboardMarkup{}

	wa := tu.WebAppInfo("https://gocha.fullioclub.ru/")

	menu.InlineKeyboard = append(menu.InlineKeyboard,
		tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("🐣 Открыть Тамагочи").WithWebApp(wa),
		),
	)

	_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Откройте интерфейс, чтобы ухаживать за питомцем!").WithReplyMarkup(menu))
	if err != nil {
		_, _ = ctx.Bot().SendMessage(ctx, tu.Messagef(tu.ID(chatID), "Ошибка!: %s", err))
	}
}
