package lua_ihui

import (
	"github.com/rverpillot/ihui"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type luaRenderer struct {
	L     *lua.LState
	table lua.LValue
	fct   lua.LValue
}

func (r *luaRenderer) Render(el *ihui.HTMLElement) error {
	if r.table != nil {
		return r.L.CallByParam(lua.P{Fn: r.fct}, r.table, luar.New(r.L, el))
	} else {
		return r.L.CallByParam(lua.P{Fn: r.fct}, luar.New(r.L, el))
	}
}

func newRenderer(L *lua.LState) int {
	arg := L.Get(1)
	LL, _ := L.NewThread()
	switch arg.Type() {
	case lua.LTTable:
		table := L.CheckTable(1)
		fct := table.RawGetString("render")
		if fct.Type() != lua.LTFunction {
			L.RaiseError("render function not found")
			return 0
		}
		renderer := &luaRenderer{
			L:     LL,
			table: table,
			fct:   fct,
		}
		L.Push(luar.New(L, renderer))

	case lua.LTFunction:
		fct := L.CheckFunction(1)
		renderer := &luaRenderer{
			L:   LL,
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
	table.RawSetString("handle", luar.New(L, ihui.Handle))
	table.RawSetString("to_renderer", L.NewFunction(newRenderer))
	L.Push(table)
	return 1
}
