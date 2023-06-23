package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// 查看修改的key存在与否，需要在所有结点查找
func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[1])
	dest := string(args[2])

	srcPeer := cluster.peerPicker.PickNode(src) //地址
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one slot in cluster mode")
	}
	return cluster.relay(srcPeer, c, args)
}
