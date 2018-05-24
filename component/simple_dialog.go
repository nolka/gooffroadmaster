package component

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
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
	s.Query("Okay! Hello there!", msg)
	s.Reg = &RegistrationInfo{}
}

func (s *HelloState) OnExit(msg *tgbotapi.Message) {
}

func (s *HelloState) Update(msg *tgbotapi.Message) {
	if s.msgCount == 0 {
		s.Query("Please, enter your first name", msg)
		s.msgCount++
		return
	}
	if s.Reg.FirstName == "" {
		s.Reg.FirstName = msg.Text
		s.Query("Please, enter Last Name!", msg)
		return
	}

	if s.Reg.LastName == "" {
		s.Reg.LastName = msg.Text
		s.Query("Okay! We are ready to enter to the world on journey! Write EnterOne to enter...", msg)
		return
	}

	if !s.Reg.Approved {
		s.Query("Type 'agree' to confirm your registration!", msg)
	}

	log.Printf("HELLO STATE: %s\n", msg)
	if strings.ToLower(msg.Text) == "agree" {
		ns := &EnterOne{s.Manager}
		s.Manager.SaveData(s.Reg)
		s.Manager.SetState(ns)
		return
	}
	log.Printf("HELO: %s\n", msg)
}

func (s *HelloState) Query(text string, prevMsg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(prevMsg.Chat.ID, text)
	s.Manager.ComponentManager.Results <- msg
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

func (s *EnterOne) Query(text string, prevMsg *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(prevMsg.Chat.ID, text)
	s.Manager.ComponentManager.Results <- msg
}
