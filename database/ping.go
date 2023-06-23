package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// 各种命令方法的具体实现
// ping服务器
func Ping(db *DB, args [][]byte) resp.Reply {
	// if len(args) == 0 {
	// 	return &reply.PongReply{}
	// } else if len(args) == 1 {
	// 	return reply.MakeStatusReply(string(args[0]))
	// } else {
	// 	return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	// }
	return reply.MakePongReply()
}

// 注册到cmdTable上进行调用
// go提供的init功能，在任何地方实现init方法，都会保证在包初始化时调用init方法
func init() {
	RegisterCommand("ping", Ping, 1)
}
