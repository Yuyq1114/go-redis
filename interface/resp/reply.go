package resp

//客户端的回复
type Reply interface {
	ToBytes() []byte //回复的内容转为字节

}
