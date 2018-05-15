package component

import (
	"gopkg.in/telegram-bot-api.v4"
)

type StateInterface interface {
	OnEnter(msg *tgbotapi.Message)
	OnExit(msg *tgbotapi.Message)
	Update(msg *tgbotapi.Message)
}

type StateManager struct {
	ComponentManager *ComponentManager
	StateStack []StateInterface
	LastMessage *tgbotapi.Message
}

func InitNewManager(componentMgr *ComponentManager, initState *StateInterface, lastMessage *tgbotapi.Message) *StateManager {
	mgr := &StateManager{}
	mgr.ComponentManager = componentMgr
	mgr.LastMessage = lastMessage
	if initState == nil {
		s := new(HelloState)
		s.Manager = mgr
		mgr.SetState(s)
	}
	return mgr
}

func (s *StateManager) GetState() StateInterface {
	return s.StateStack[len(s.StateStack)-1]
}

func (s *StateManager) GetPrevState() *StateInterface {
	return &s.StateStack[len(s.StateStack)-2]
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

/**
	Base state
 */
type State struct {
}

type UpdateArgs struct {

}

func (s *State) Update(args UpdateArgs) {
}
