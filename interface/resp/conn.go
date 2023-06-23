package resp

//resp表示redis的协议
//在redis协议层，代表redis连接
type Connection interface {
	Write([]byte) error //给客户端回复消息
	GetDBIndex() int    //获取DB索引
	SelectDB(int)       //切换DB
}
