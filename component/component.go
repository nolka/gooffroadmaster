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
	Init(manager *ComponentManager)
	HandleMessage(update tgbotapi.Update)
	HandleCallback(update tgbotapi.Update)
}

func NewComponentManager(bot *tgbotapi.BotAPI, results chan tgbotapi.MessageConfig) *ComponentManager {
	cm := &ComponentManager{}
	cm.Bot = bot
	cm.Results = results
	cm.Components = make(map[int]BotMessageComponentInterface)
	return cm
}

type ComponentManager struct {
	Bot        *tgbotapi.BotAPI
	Results    chan tgbotapi.MessageConfig
	Components map[int]BotMessageComponentInterface
}

func (m *ComponentManager) GetComponents() map[int]BotMessageComponentInterface {
	return m.Components
}

func (m *ComponentManager) RegisterComponent(component BotMessageComponentInterface) {
	component.SetId(len(m.Components))
	m.Components[len(m.Components)] = component
}

func (m *ComponentManager) Dispatch(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		log.Println("Dispatching callback...")
		parts := strings.SplitN(update.CallbackQuery.Data, "|", 1)
		componentId, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Failed to get component id from string: %s. Skipping...\n", parts[0])
		}
		components := m.GetComponents()
		components[componentId].HandleCallback(update)
		return
	}

	log.Println("Dispatching message...")
	for _, c := range m.GetComponents() {
		c.HandleMessage(update)
	}
}

func (m *ComponentManager) Halt() {
	for _, c := range m.GetComponents() {
		util.SaveConfig(c)
	}
}
