package component

import (
	"fmt"
	"log"
	"strings"

	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

func NewInteractiveMenu(manager *ComponentManager) *InteractiveMenu {
	c := &InteractiveMenu{}
	util.LoadConfig(c)
	c.Init(manager)
	return c
}

type InteractiveMenu struct {
	Manager  *ComponentManager           `json:"-"`
	Id       int                         `json:"-"`
	UserList map[int]*StateManager `json:"-"`
}

func (i *InteractiveMenu) SetId(id int) {
	i.Id = id
}

func (i *InteractiveMenu) Init(manager *ComponentManager) {
	i.Manager = manager
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
	_, ok := i.UserList[userId]
	if !ok {
		log.Printf("Creating state mgr for user: %s\n", userId)
		i.UserList[userId] = InitNewManager(i.Manager, nil, update.Message)
	}
	s := i.UserList[userId]
	log.Printf("Dispatching state message to user id: %d\n", userId)
	s.Update(update.Message)
}

func (i *InteractiveMenu) HandleCallback(update tgbotapi.Update) {

}
