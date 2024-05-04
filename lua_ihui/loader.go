package lua_ihui

import (
	"github.com/rverpillot/ihui"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func Loader(L *lua.LState) int {
	table := L.NewTable()
	table.RawSetString("handle", luar.New(L, ihui.Handle))
	L.Push(table)
	return 1
}
