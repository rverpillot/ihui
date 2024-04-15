package ihui

type Event struct {
	Name    string
	Id      string
	Target  string
	Refresh bool
	Data    interface{}
}

func (e *Event) Value() string {
	return e.Data.(string)
}

func (e *Event) IsChecked() bool {
	t, ok := e.Data.(bool)
	if !ok {
		return false
	}
	return t
}
