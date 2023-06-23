package cluster

//cluster 集群层 转发层
import (
	"context"
	"fmt"
	"go-redis/config"
	"go-redis/database"
	databaseface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"runtime/debug"
	"strings"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type ClusterDatabase struct {
	self           string                      //自己的名称地址
	nodes          []string                    //集群结点切片
	peerPicker     *consistenthash.NodeMap     //一致性hash管理器
	peerConnection map[string]*pool.ObjectPool //连接池。string存储结点地址
	db             databaseface.Database       //数据库
}

// new一个ClusterDatabase
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self: config.Properties.Self, //配置文件中的自己的地址

		db:             database.NewStandaloneDatabase(), //每个结点调用单机版本
		peerPicker:     consistenthash.NewNodeMap(nil),   //传入函数为nil自动使用CRC
		peerConnection: make(map[string]*pool.ObjectPool),
	} //初始化clusterdB

	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	// for _, peer := range config.Properties.Peers {
	// 	nodes = append(nodes, peer)
	// }
	// node中存入配置的连接地址
	nodes = append(nodes, config.Properties.Peers...)
	nodes = append(nodes, config.Properties.Self)
	//一致性hash的环
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer, //连接池自动维护连接，默认8个
		})
	}
	cluster.nodes = nodes
	return cluster
}

// 根据不同的指令调用不同的执行模式
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply

// 关闭cluster也就是关闭数据库
func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

var router = makeRouter()

// 执行
func (cluster *ClusterDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		//recover
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, cmdLine)
	return
}

// 关闭后操作
func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
