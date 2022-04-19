package actorsystem

import "fmt"

type statistics struct {
	Name             string         `yaml:"name,omitempty"`
	MessagesByType   map[string]int `yaml:"message_by_type,omitempty"`
	MessagesPosted   int            `yaml:"messages_posted,omitempty"`
	MessagesReceived int            `yaml:"messages_received,omitempty"`
}

func (s *statistics) MailboxStarted() {

}

func (s *statistics) MessagePosted(message interface{}) {
	s.MessagesPosted++
	t := fmt.Sprintf("%T", message)
	if t == "string" {
		t = message.(string)
	}
	if _, found := s.MessagesByType[t]; !found {
		s.MessagesByType[t] = 0
	}
	s.MessagesByType[t]++
}

func (s *statistics) MessageReceived(message interface{}) {
	s.MessagesReceived++
}

func (s *statistics) MailboxEmpty() {

}
