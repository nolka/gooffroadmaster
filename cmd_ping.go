package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"os/exec"
)

type Ping struct {
	Command
}

func (p *Ping) Handle(message *tgbotapi.Message, bot *tgbotapi.BotAPI) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.ReplyToMessageID = message.MessageID
	if len(p.Args) == 0 {
		msg.Text = "No required params set"
		return msg, nil
	}

	host := p.Args[0]
	pingCmd, _ := exec.LookPath("ping")
	result, err := exec.Command(pingCmd, host, "-c 4").Output()
	if err != nil {
		msg.Text = err.Error()
		return msg, nil
	}

	msg.Text = string(result)
	return msg, nil
}
