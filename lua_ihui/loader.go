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
	LL, _ := r.L.NewThread()
	if r.table != nil {
		return LL.CallByParam(lua.P{Fn: r.fct}, r.table, luar.New(LL, el))
	} else {
		return LL.CallByParam(lua.P{Fn: r.fct}, luar.New(LL, el))
	}
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
			L:     L,
			table: table,
			fct:   fct,
		}
		L.Push(luar.New(L, renderer))

	case lua.LTFunction:
		fct := L.CheckFunction(1)
		renderer := &luaRenderer{
			L:   L,
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
