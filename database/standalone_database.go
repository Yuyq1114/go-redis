package database

import (
	"fmt"
	"go-redis/aof"
	"go-redis/config"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// 实现数据库核心
type StandaloneDatabase struct {
	dbSet      []*DB           //持有一组分db的指针
	aofHandler *aof.AofHandler //aof文件
}

// 新建数据库，初始化一个redis内核，内部有n个分数据库
func NewStandaloneDatabase() *StandaloneDatabase {
	database := &StandaloneDatabase{}
	if config.Properties.Databases == 0 { //获取config文件中的databases字段大小
		config.Properties.Databases = 16 //默认16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet { //循环初始化分db
		db := makeDB()
		db.index = i
		database.dbSet[i] = db
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAOFHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			//for循环让db.index始终等于15
			//逃逸到了堆上
			sdb := db
			sdb.addAof = func(line CmdLine) {
				database.aofHandler.AddAof(sdb.index, line)
			}
		}
	}

	return database
}

// 执行方法|set k v|get k|select 2
// 把用户输入的指令抓发给分db，调用分db的Exec方法
func (mdb *StandaloneDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {

	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" {
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}
	// normal commands
	dbIndex := c.GetDBIndex()
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	selectedDB := mdb.dbSet[dbIndex]
	return selectedDB.Exec(c, cmdLine)
}

// Close graceful shutdown database
func (mdb *StandaloneDatabase) Close() {
}

func (mdb *StandaloneDatabase) AfterClientClose(c resp.Connection) {
}

// 选择用户执行的db
// 通过用户发送的指令修改connection中的字段|select 2
func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex > len(database.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
