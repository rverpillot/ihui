package lua_ihui

import (
	"github.com/rverpillot/ihui"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type luaRenderer struct {
	table lua.LValue
	fct   lua.LValue
}

func (r *luaRenderer) Render(el *ihui.HTMLElement) error {
	L := el.Session().Get("_lua").(*lua.LState)
	if r.table != nil {
		return L.CallByParam(lua.P{Fn: r.fct}, r.table, luar.New(L, el))
	} else {
		return L.CallByParam(lua.P{Fn: r.fct}, luar.New(L, el))
	}
}

func handle(L *lua.LState) int {
	contextRoot := L.CheckString(1)
	fct := L.CheckFunction(2)
	ihui.Handle(contextRoot, func(s *ihui.Session) error {
		LL, _ := L.NewThread() //TODO: do we need to close this thread?
		s.Set("_lua", LL)
		return L.CallByParam(lua.P{Fn: fct}, luar.New(LL, s))
	})
	return 0
}

func newRenderer(L *lua.LState) int {
	arg := L.Get(1)
	switch arg.Type() {
	case lua.LTTable:
		table := L.CheckTable(1)
		fct := table.RawGetString("render")
		if fct.Type() != lua.LTFunction {
			L.RaiseError("render function not found")
			return 0
		}
		renderer := &luaRenderer{
			table: table,
			fct:   fct,
		}
		L.Push(luar.New(L, renderer))

	case lua.LTFunction:
		fct := L.CheckFunction(1)
		renderer := &luaRenderer{
			fct: fct,
		}
		L.Push(luar.New(L, renderer))

	default:
		L.RaiseError("invalid argument")
	}

	return 1
}

func Loader(L *lua.LState) int {
	table := L.NewTable()
	table.RawSetString("handle", L.NewFunction(handle))
	table.RawSetString("to_renderer", L.NewFunction(newRenderer))
	L.Push(table)
	return 1
}
