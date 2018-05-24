package controllers

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"strconv"
	"strings"
)

type HelloState struct {
	Manager   *StateManager
	Reg *RegistrationInfo
	LastName  string
	msgCount  int
}

type RegistrationInfo struct {
	FirstName string
	LastName string
	Approved bool
}

func (s *HelloState) OnEnter(msg *tgbotapi.Message) {
	if msg.Chat.Type != "private"{
		return;
	}
	s.Say("Okay! Hello there!", msg)
	s.Reg = &RegistrationInfo{}
}

func (s *HelloState) OnExit(msg *tgbotapi.Message) {
}

func (s *HelloState) PrepareData(d ...string) string {
	return fmt.Sprintf("%d|%s", s.Manager.Menu.Id, strings.Join(d, "|"))
}


func (s *HelloState) Update(msg *tgbotapi.Message) {
	if s.msgCount == 0 {
		s.Say("Please, enter your first name", msg)
		s.msgCount++
		return
	}
	if s.Reg.FirstName == "" {
		s.Reg.FirstName = msg.Text
		s.Say("Please, enter Last Name!", msg)
		return
	}

	if s.Reg.LastName == "" {
		s.Reg.LastName = msg.Text
	}

	if !s.Reg.Approved {
		s.QueryConfirmation(msg)
	}

	log.Printf("HELO: %s\n", msg)
}

func (s *HelloState) QueryConfirmation(msg *tgbotapi.Message) {
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Confirm", s.PrepareData(strconv.Itoa(msg.From.ID), "yes")),
			tgbotapi.NewInlineKeyboardButtonData("Cancel", s.PrepareData(strconv.Itoa(msg.From.ID), "no")),
		),
	)
	c := tgbotapi.NewMessage(msg.Chat.ID, "")
	c.ReplyToMessageID = msg.MessageID
	c.ReplyMarkup = markup
	c.Text = "Confirm registration"
	s.Manager.Menu.Router.Results <- c
}

func (s *HelloState) UpdateCallback(msg *tgbotapi.CallbackQuery, userId int) {
	parts := strings.Split(msg.Data, "|")

	if parts[2] == "yes" {
		ns := &EnterOne{s.Manager}
		s.Manager.SaveData(s.Reg)
		s.Manager.SetState(ns)
		return
	} else {
		s.Say("Registration aborted", msg.Message)
	}
}

func (s *HelloState) Say(text string, prevMsg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(prevMsg.Chat.ID, text)
	s.Manager.Menu.Router.Results <- msg
}

type EnterOne struct {
	Manager *StateManager
}

type HelloStateFactory func() StateInterface

func (s *EnterOne) OnEnter(msg *tgbotapi.Message) {
	var reg *RegistrationInfo = s.Manager.GetSavedData().(*RegistrationInfo)

	s.Query(fmt.Sprintf("%s %s, You are in EnterOne journey world!!", reg.FirstName, reg.LastName), msg)
}

func (s *EnterOne) OnExit(msg *tgbotapi.Message) {
	s.Query("Bye! See ya later!", msg)
}

func (s *EnterOne) Update(msg *tgbotapi.Message) {
	if msg.Text == "one" {
		s.Manager.PopState()
		return
	}
}

func (s *EnterOne) UpdateCallback(msg *tgbotapi.CallbackQuery, userId int) {

}

func (s *EnterOne) Query(text string, prevMsg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(prevMsg.Chat.ID, text)
	s.Manager.Menu.Router.Results <- msg
}
