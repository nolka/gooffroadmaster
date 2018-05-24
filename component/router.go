package component

import (
	"log"
	"strconv"
	"strings"

	"github.com/nolka/gooffroadmaster/util"
	"gopkg.in/telegram-bot-api.v4"
)

type BotMessageComponentInterface interface {
	SetId(id int)
	GetName() string
	Init(manager *Router)
	HandleMessage(update tgbotapi.Update)
	HandleCallback(update tgbotapi.Update)
}

func NewMessageRouter(bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) *Router {
	cm := &Router{}
	cm.Bot = bot
	cm.Results = results
	cm.Controllers = make(map[int]BotMessageComponentInterface)
	return cm
}

type Router struct {
	Bot         *tgbotapi.BotAPI
	Results     chan tgbotapi.MessageConfig
	Controllers map[int]BotMessageComponentInterface
}

func (m *Router) GetControllers() map[int]BotMessageComponentInterface {
	return m.Controllers
}

func (m *Router) RegisterController(component BotMessageComponentInterface) {
	component.SetId(len(m.Controllers))
	m.Controllers[len(m.Controllers)] = component
}

func (m *Router) Dispatch(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		parts := strings.SplitN(update.CallbackQuery.Data, "|", 2)
		componentId, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Failed to get component id from string: %s. Skipping...\n", parts[0])
			return
		}
		components := m.GetControllers()
		components[componentId].HandleCallback(update)
		return
	}

	for _, c := range m.GetControllers() {
		c.HandleMessage(update)
	}
}

func (m *Router) Halt() {
	for _, c := range m.GetControllers() {
		util.SaveConfig(c)
	}
}
