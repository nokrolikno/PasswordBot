package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nokrolikno/PasswordBot/internal/credentials"
	"github.com/nokrolikno/PasswordBot/internal/handlers"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Token string `yaml:"token"`
}

func deleteMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	timer := time.NewTimer(time.Second * 10)
	<-timer.C
	bot.Send(
		tgbotapi.NewDeleteMessage(chatID, messageID),
	)
}

func startPasswordBot(cfg Config, ctx context.Context) error {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := zapLogger.Sync() // flushes buffer, if any
		if err != nil {
			panic(err)
		}
	}()
	logger := zapLogger.Sugar()

	credentialsRepo := credentials.NewMemoryRepo()
	credentialsHandler := handlers.CredentialsHandler{CredentialsRepo: credentialsRepo, Logger: logger}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return fmt.Errorf("NewBotAPI failed: %w", err)
	}
	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(cfg.Server.Host)
	if err != nil {
		return fmt.Errorf("NewWebhook failed: %w", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		return fmt.Errorf("SetWebhook failed: %w", err)
	}

	updates := bot.ListenForWebhook("/")

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all is working"))
	})

	port := cfg.Server.Port
	go func() {
		logger.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)

	handleCommands := make(map[string]func(userID int64, commandText string) (string, error))
	handleCommands["/start"] = credentialsHandler.Start
	handleCommands["/set"] = credentialsHandler.Set
	handleCommands["/get"] = credentialsHandler.Get
	handleCommands["/del"] = credentialsHandler.Del
	handleCommands["/help"] = credentialsHandler.Help
	handleCommands["/getServices"] = credentialsHandler.GetServices

	var helpKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Помощь"),
		),
	)

	for update := range updates {
		logger.Info("upd: %#v\n", update)
		if update.Message == nil {
			continue
		}
		command := strings.SplitN(update.Message.Text, " ", 2)
		if len(command) < 2 {
			command = append(command, "")
		}
		function, ok := handleCommands[command[0]]
		if !ok {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Неизвестная команда",
			)
			msg.ReplyMarkup = helpKeyboard
			bot.Send(msg)
			continue
		}
		response, err := function(update.Message.From.ID, command[1])
		if errors.Is(err, handlers.ErrHandler) {
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Что-то пошло не так, повторите попытку позже",
			)
			msg.ReplyMarkup = helpKeyboard
			bot.Send(msg)
			continue
		}
		if command[0] == "/set" && err == nil {
			go deleteMessage(bot, update.Message.Chat.ID, update.Message.MessageID)
		}
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			response,
		)
		msg.ReplyMarkup = helpKeyboard
		message, _ := bot.Send(msg)
		if command[0] == "/get" && err == nil {
			go deleteMessage(bot, update.Message.Chat.ID, message.MessageID)
		}
	}
	return nil
}

func main() {
	f, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	err = startPasswordBot(cfg, context.Background())
	if err != nil {
		panic(err)
	}
}
