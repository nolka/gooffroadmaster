package component

import (
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"strings"
)

type HelloState struct {
	Manager   *StateManager
	FirstName string
	LastName  string
	msgCount  int
}

func (s *HelloState) OnEnter(msg *tgbotapi.Message) {
	s.Query("Okay! Hello there!", msg)
}

func (s *HelloState) OnExit(msg *tgbotapi.Message) {
}

func (s *HelloState) Update(msg *tgbotapi.Message) {
	if s.msgCount == 0 {
		s.Query("Please, enter your first name", msg)
		s.msgCount++
		return
	}
	if s.FirstName == "" {
		s.FirstName = msg.Text
		s.Query("Please, enter Last Name!", msg)
		return
	}

	if s.LastName == "" {
		s.LastName = msg.Text
		s.Query("Okay! We are ready to enter to the world on journey! Write EnterOne to enter...", msg)
		return
	}

	log.Printf("HELLO STATE: %s\n", msg)
	if strings.ToLower(msg.Text) == "enterone" {
		ns := &EnterOne{s.Manager}
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

func (s *EnterOne) OnEnter(msg *tgbotapi.Message) {
	var prev HelloState = &s.Manager.GetPrevState()

	s.Query(fmt.Sprintf("%s %s, You are in EnterOne journey world!!", prev.(HelloState).FirstName, prev.(HelloState).LastName), msg)
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
