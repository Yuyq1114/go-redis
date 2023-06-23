package aof

import (
	"go-redis/config"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
)

// redis持久化，AOF保存写命令
type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16 //payload长度
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler 从管道接收msg写入到aof文件中
type AofHandler struct {
	db          databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int
}

// 新建一个AOF写命令
func NewAOFHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename //获取文件名
	handler.db = db                                        //获取数据库
	// handler.LoadAof() //恢复之前的文件，写入
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile                           //填充aof文件
	handler.aofChan = make(chan *payload, aofQueueSize) //从channel中取出条目
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof 通过管道发送命令到aof goroutine
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	// 判断aof功能开没开，config文件中的appendonly条目
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// 从channel中不断取，写入文件
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { //每次检测dbIndex和上次是否一样
			//不一样插入select语句
			//utils.ToCmdLine编码成二维字节切片
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue
			}
			handler.currentDB = p.dbIndex //保存当前DB信息，不用每次都添加DBindex
		}
		//一样就直接插入语句
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

// 加载aof文件
func (handler *AofHandler) LoadAof() {

	file, err := os.Open(handler.aofFilename) //加载aof文件
	if err != nil {
		logger.Warn(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)       //解析文件命令
	fakeConn := &connection.Connection{} // 存储dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args) //执行命令
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", err)
		}
	}
}
