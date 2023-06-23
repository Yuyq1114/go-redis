package database

import "go-redis/interface/resp"

type CmdLine = [][]byte //别名 二维字节组

//代表redis层的业务核心
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply //实现执行
	Close()
	AfterClientClose(c resp.Connection) //关闭后的一些善后功能
}

//空接口指代redis数据类型，如string,list ,set
type DataEntity struct {
	Data interface{}
}
