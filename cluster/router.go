package cluster

import "go-redis/interface/resp"

type CmdLine = [][]byte

// 路由
// routerMap中存放指令对应的方法
func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	routerMap["ping"] = ping

	routerMap["del"] = Del

	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename

	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc

	routerMap["flushdb"] = FlushDB
	routerMap["select"] = execSelect

	return routerMap
}

// 默认方法，转发
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	key := string(args[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, args)
}
