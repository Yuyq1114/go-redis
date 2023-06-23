package handler

import (
	//前面的字是起一个别名
	"context"
	"go-redis/cluster"
	"go-redis/config"
	"go-redis/database"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"strings"

	"net"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknow\r\n")
)

type RespHandler struct {
	activeConn sync.Map              //记录所有客户端的连接chan
	db         databaseface.Database //持有一个redis核心的业务层
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseface.Database = database.NewStandaloneDatabase()
	//如果配置文件有，则使用集群版本
	if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	}

	return &RespHandler{
		db: db,
	}
}

// 关闭一个连接
func (r *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)

}

// 处理tcp的连接
func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
	}
	client := connection.NewConn(conn)
	r.activeConn.Store(client, struct{}{}) //保存连接的客户端信息
	ch := parser.ParseStream(conn)         //信息让Parse进行解析
	for payload := range ch {
		//如果有错误
		if payload.Err != nil {

			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				//说明客户端和我们挥手，关闭连接
				r.closeClient(client)
				logger.Info("connection closed:" + client.RemoteAddr().String())
				return
			}
			//协议错误
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("connection closed:" + client.RemoteAddr().String())
				return
			}
			continue

		}
		//无错误则exec执行
		if payload.Data == nil {
			continue
		}
		reply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := r.db.Exec(client, reply.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}

}

// 关闭整个client，所有连接
func (r *RespHandler) Close() error {
	logger.Info("handler shutting down")
	r.closing.Set(true)
	r.activeConn.Range(
		func(key interface{}, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		})
	r.db.Close()
	return nil
}
