package database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

// 实现 DEL
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted)) //给用户发送的返回值
}

// EXISTS
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLine2("flushdb", args...))
	return reply.MakeOkReply()
}

// 回复键值类型
// TYPE
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exises := db.GetEntity(key)
	if !exises {
		reply.MakeStatusReply("none") //TCP :none\r\n
	}
	switch entity.Data.(type) {
	case []byte:
		reply.MakeStatusReply("string")

	}
	return &reply.UnknownErrReply{}
}

// RENAME将key改名,底层上直接删除
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("rename", args...))
	return reply.MakeOkReply()

}

// RENAMENX 检查REname时是否会覆盖存在
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest) //判断是否存在
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(src)
	if !exists {
		reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLine2("renamenx", args...))

	return reply.MakeIntReply(1)

}

// KEYS * 列出所有符合条件的key
// 通配符lib\wildcard\wildcard.go
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0) //存放所有符合条件的key
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

// 初始化,注册进cmd
func init() {
	RegisterCommand("DEL", execDel, -2) //-2表示变长，自己定义的
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("FLUSHDB", execFlushDB, -1)
	RegisterCommand("TYPE", execType, 2)
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenamenx, 3)
	RegisterCommand("KEYS", execKeys, 2) //keys *
}
