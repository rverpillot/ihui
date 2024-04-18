package ihui

import (
	"fmt"
)

type ActionCallback func(*Session, Event) error

func (f ActionCallback) String() string { return fmt.Sprintf("#%p", f) }

type Action struct {
	Selector string
	Name     string
	Fct      ActionCallback
}
