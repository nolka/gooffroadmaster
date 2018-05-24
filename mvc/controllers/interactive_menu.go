package controllers

import (
	"fmt"
	"github.com/nolka/gooffroadmaster/mvc"
	"log"
	"strings"

	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

func NewInteractiveMenu(manager *mvc.Router) *InteractiveMenu {
	c := &InteractiveMenu{}
	util.LoadConfig(c)
	c.Init(manager)
	return c
}

type InteractiveMenu struct {
	Router   *mvc.Router           `json:"-"`
	Id       int                   `json:"-"`
	UserList map[int]*StateManager `json:"-"`
}

func (i *InteractiveMenu) SetId(id int) {
	i.Id = id
}

func (i *InteractiveMenu) Init(manager *mvc.Router) {
	i.Router = manager
	i.UserList = make(map[int]*StateManager)
}

func (i *InteractiveMenu) GetName() string {
	return "Interactive menu example"
}

func (i *InteractiveMenu) PrepareData(d ...string) string {
	return fmt.Sprintf("%d|%s", i.Id, strings.Join(d, "|"))
}

func (i *InteractiveMenu) HandleMessage(update tgbotapi.Update) {
	if update.Message.Chat.Type != "private" {
		return;
	}
	userId := update.Message.From.ID
	s, ok := i.UserList[userId]
	if !ok {
		log.Printf("Creating state mgr for user: %d\n", userId)
		s = InitNewManager(i, nil, update.Message)
		i.UserList[userId] = s
	}
	log.Printf("Dispatching state message to user id: %d\n", userId)
	s.Update(update.Message)
}

func (i *InteractiveMenu) HandleCallback(update tgbotapi.Update) {
	userId := update.CallbackQuery.From.ID
	s, ok := i.UserList[userId]
	if !ok {
		log.Printf("Error handling callback because user id: %d was not requested this action!\n", userId)
		return
	}
	s.UpdateCallback(update.CallbackQuery, userId)
}
