package main

import (
	"errors"
	"gopkg.in/telegram-bot-api.v4"
)

type Command struct {
	Command string
	Args    []string
}

type ExecutableCommand interface {
	Handle(message *tgbotapi.Message, bot *tgbotapi.BotAPI) (tgbotapi.MessageConfig, error)
	SetArgs(args []string)
}

func (c *Command) SetArgs(args []string) {
	c.Args = args
}

func (c *Command) Handle(message *tgbotapi.Message, bot *tgbotapi.BotAPI) (tgbotapi.MessageConfig, error) {
	return tgbotapi.MessageConfig{}, errors.New("Not implemented!")
}

func EnumerateCommands() map[string]ExecutableCommand {
	return map[string]ExecutableCommand{
		"ping": &Ping{},
	}
}
