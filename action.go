package ihui

import (
	"fmt"
)

type ActionFunc func(*Session, Event) bool

func (f ActionFunc) String() string { return fmt.Sprintf("#%p", f) }

type Action struct {
	Selector string
	Name     string
	Fct      ActionFunc
}
