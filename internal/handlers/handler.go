package handlers

import (
	"fmt"
	"strings"

	"github.com/nokrolikno/PasswordBot/internal/credentials"
	"go.uber.org/zap"
)

type CredentialsHandler struct {
	CredentialsRepo credentials.CredentialsRepo
	Logger          *zap.SugaredLogger
}

func (ch *CredentialsHandler) Start(userID int64, commandText string) (string, error) {
	return `Привет! Я бот для хранения твоих паролей для любых сервисов

/set - запомнить логин и пароль для сервиса
Использование:
/set <сервис> <логин> <пароль>

/get - напомнить пароль
Использование:
/get <сервис>

/del - забыть пароль
Использование:
/del <сервис>
`, nil
}

func (ch *CredentialsHandler) Set(userID int64, commandText string) (string, error) {
	words := strings.Split(commandText, " ")
	if len(words) != 3 {
		return "Использование:\n/set <сервис> <логин> <пароль>\nНапример:\n/set telegram qwerty 12345", nil
	}
	service := words[0]
	login := words[1]
	password := words[2]
	err := ch.CredentialsRepo.Set(userID, service, login, password)
	if err != nil {
		return "", fmt.Errorf("error in CredentialsRepo.Set: %w", err)
	}
	return fmt.Sprintf(`Логин и пароль для сервиса "%s" установлены`, service), nil
}

func (ch *CredentialsHandler) Get(userID int64, commandText string) (string, error) {
	if len(commandText) == 0 {
		return "Использование:\n/get <сервис>\nНапример:\n/get telegram", nil
	}
	service := commandText
	login, password, ok, err := ch.CredentialsRepo.Get(userID, service)
	if err != nil {
		return "", fmt.Errorf("error in CredentialsRepo.Get: %w", err)
	}
	if !ok {
		return fmt.Sprintf("Логин и пароль для cервиса \"%s\" я ещё не запоминал\nПопробуй команду /set", service), nil
	}
	return fmt.Sprintf("Логин и пароль для сервиса \"%s\":\nЛогин: %s\nПароль: %s", service, login, password), nil
}

func (ch *CredentialsHandler) Del(userID int64, commandText string) (string, error) {
	words := strings.Split(commandText, " ")
	if len(words) != 1 {
		return "Использование:\n/del <сервис>\nНапример:\n/del telegram", nil
	}
	service := words[0]
	ok, err := ch.CredentialsRepo.Del(userID, service)
	if err != nil {
		return "", fmt.Errorf("error in CredentialsRepo.Del: %w", err)
	}
	if ok {
		return fmt.Sprintf("Логин и пароль для сервиса \"%s\" сброшены", service), nil
	}
	return fmt.Sprintf("Сервис \"%s\" не найден\nНечего сбрасывать", service), nil
}