package database

import "strings"

//每一个指令对应一个command结构体
var cmdTable = make(map[string]*command)

// 每一个command结构体里有一个执行方法，每一个指令都是一个command结构体，
// 在DB里把方法施加到DB上
type command struct {
	exextor ExecFunc //执行方法，ExecFunc是一个方法
	arity   int      //执行参数
}

//注册一些指令的实现，输入方法就可以自动将方法新建command结构体，放入cmdTable中
func RegisterCommand(name string, exector ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		exextor: exector,
		arity:   arity,
	}
}
