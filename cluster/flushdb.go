package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// flush需要广播
func FlushDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply reply.ErrorReply
	for _, v := range replies { //所有的结点都ok才可以
		if reply.IsErrorReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return &reply.OkReply{}
	}
	return reply.MakeErrReply("error occurs: " + errReply.Error())
}
