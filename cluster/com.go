package cluster

import (
	"context"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/client"
	"go-redis/resp/reply"
	"strconv"
)

// 负责与其他结点的通信

// 获取连接客户端
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	factory, ok := cluster.peerConnection[peer] //获取当前结点连接池
	if !ok {
		return nil, errors.New("connection factory not found")
	}
	raw, err := factory.BorrowObject(context.Background()) //借一个连接
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(*client.Client) //转换
	if !ok {
		return nil, errors.New("connection factory make wrong type")
	}
	return conn, nil
}

// 还连接，防止连接池耗尽
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}

// 实现转发模式
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self { //如果结点时自己
		// to self db
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer) //获取连接池节点
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	//底层实现select命令
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	return peerClient.Send(args) //发送命令
}

// 实现群发模式
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	result := make(map[string]resp.Reply)
	//遍历结点，每一个都转发，实现群发
	for _, node := range cluster.nodes {
		reply := cluster.relay(node, c, args)
		result[node] = reply
	}
	return result
}
