package database

import (
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
// 	entity, ok := db.GetEntity(key)
// 	if !ok {
// 		return nil, nil
// 	}
// 	bytes, ok := entity.Data.([]byte)
// 	if !ok {
// 		return nil, &reply.WrongTypeErrReply{}
// 	}
// 	return bytes, nil
// }

// GET 获取键的值
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	// bytes, err := db.getAsString(key)
	entity, exists := db.GetEntity(key) //获得key
	if !exists {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes) //存在转化为BulkReply并返回
}

// SET KEY V
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity) //包装成DataEntity并Put
	// db.addAof(args)
	db.addAof(utils.ToCmdLine2("set", args...))
	return &reply.OkReply{}
}

// SETNX 如果不存在则SET
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.addAof(utils.ToCmdLine2("setnx", args...))
	result := db.PutIfAbsent(key, entity)
	return reply.MakeIntReply(int64(result))
}

// GETSET 先获取key1的值，然后set成新值，返回原来的值
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]

	entity, exists := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{Data: value})
	db.addAof(utils.ToCmdLine2("getset", args...))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	old := entity.Data.([]byte)
	return reply.MakeBulkReply(old)
}

// STRLEN key对应的value长度
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	old := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(old)))
}

// 初始化
func init() {
	RegisterCommand("Get", execGet, 2)
	RegisterCommand("Set", execSet, -3)
	RegisterCommand("SetNx", execSetNX, 3)
	RegisterCommand("GetSet", execGetSet, 3)
	RegisterCommand("StrLen", execStrLen, 2)
}
