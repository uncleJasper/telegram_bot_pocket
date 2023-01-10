package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zhashkevych/go-pocket-sdk"
	"log"
	"telegram-bot-pocket/pkg/config"
	"telegram-bot-pocket/pkg/repository"
)

type Bot struct {
	bot             *tgbotapi.BotAPI
	pocketClient    *pocket.Client
	tokenRepository repository.TokenRepository
	redirectURL     string
	messages        config.Messages
}

func NewBot(bot *tgbotapi.BotAPI, pocketClient *pocket.Client, tr repository.TokenRepository, redirectURL string, messages config.Messages) *Bot {
	return &Bot{
		bot:             bot,
		pocketClient:    pocketClient,
		tokenRepository: tr,
		redirectURL:     redirectURL,
		messages:        messages,
	}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	updates, err := b.initUpdatesChanel()
	if err != nil {
		return err
	}

	b.handleUpdates(updates)
	return nil
}

// initUpdatesChanel инициализация нового канала
func (b *Bot) initUpdatesChanel() (tgbotapi.UpdatesChannel, error) {
	// новая конфигурация с указанием периодичности запросов к серверу телеграм
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	// чтение обновления
	for update := range updates {
		if update.Message == nil { // If we got a message
			continue
		}

		//проверяем, пришла команда или нет
		if update.Message.IsCommand() {
			if err := b.handleCommand(update.Message); err != nil {
				b.handleError(update.Message.Chat.ID, err)
			}
			continue
		}

		if err := b.handleMessage(update.Message); err != nil {
			b.handleError(update.Message.Chat.ID, err)
		}
	}
}
