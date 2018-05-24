package component

import (
	"gopkg.in/telegram-bot-api.v4"
)

type StateInterface interface {
	OnEnter(msg *tgbotapi.Message)
	OnExit(msg *tgbotapi.Message)
	Update(msg *tgbotapi.Message)
	UpdateCallback(callback *tgbotapi.CallbackQuery, userId int)
}

type StateManager struct {
	Menu *InteractiveMenu
	StateStack []StateInterface
	LastMessage *tgbotapi.Message
	SavedData interface{}
}

func InitNewManager(menu *InteractiveMenu, initState *StateInterface, lastMessage *tgbotapi.Message) *StateManager {
	mgr := &StateManager{}
	mgr.Menu = menu
	mgr.LastMessage = lastMessage
	if initState == nil {
		s := new(HelloState)
		s.Manager = mgr
		mgr.SetState(s)
	}
	return mgr
}

func(s *StateManager) SaveData(data interface{}) {
	s.SavedData = data
}

func(s *StateManager) GetSavedData() interface{} {
	return s.SavedData
}

func (s *StateManager) GetState() StateInterface {
	return s.StateStack[len(s.StateStack)-1]
}

func (s *StateManager) SetState(state StateInterface) {
	state.OnEnter(s.LastMessage)
	s.StateStack = append(s.StateStack, state)
}

func (s *StateManager) PopState() StateInterface {
	si := s.StateStack[len(s.StateStack)-1]
	si.OnExit(s.LastMessage)
	s.StateStack = s.StateStack[:len(s.StateStack)-1]
	return si
}

func (s *StateManager) Update(msg *tgbotapi.Message) {
	s.LastMessage = msg
	s.GetState().Update(msg)
}

func (s *StateManager) UpdateCallback(msg *tgbotapi.CallbackQuery, userId int) {
	s.GetState().UpdateCallback(msg, userId)
}