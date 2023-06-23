package database

import (
	"go-redis/datastruct/dict"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"

	"go-redis/interface/database"
)

// redis的一个db，分数据库
type DB struct {
	index  int
	data   dict.Dict
	addAof func(CmdLine) //给DB结构一个方法，后面把Aof方法赋给这个方法，可以调用了
}

// 对所有指令的实现
type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func makeDB() *DB {
	db := &DB{
		data:   dict.MakeSyncDict(),
		addAof: func(line CmdLine) {},
	}
	return db
}

// 对指令进行实现，并返回结果，不是具体指令方法的实现
func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	//弄清用户发送的指令名，第一个切片
	cmdName := strings.ToLower(string(cmdLine[0])) //转成小写
	cmd, ok := cmdTable[cmdName]
	//没有这个命令
	if !ok {
		return reply.MakeErrReply("ERR unknown command" + cmdName)
	}
	//如果参数个数不对
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.exextor
	//指令的后面执行
	return fun(db, cmdLine[1:])
}

// 看用户输入参数个数是否合法
// 如果指令个数定长，arity = 长度
// 如果不定长，arity = -最大长度
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity

}

// 不是直接获取key，在DB层面获取key的逻辑
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	//
	raw, ok := db.data.Get(key) //调用底层Get的实现
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// put一个实体进入DB
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Remove多个key
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Clear
func (db *DB) Flush() {
	db.data.Clear()

}
