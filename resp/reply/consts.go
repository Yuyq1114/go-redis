package reply

//保存一些固定的回复

//1 回复PONG
type PongReply struct {
}

var pongbytes = []byte("+PONG\r\n")

func (p PongReply) ToBytes() []byte {
	return pongbytes
}

//编程习惯种暴露一个make方法便于外部的修改
func MakePongReply() *PongReply {
	return &PongReply{}
}

// 2 回复OK
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes marshal redis.Reply
func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply) //在本地持有一个OkReply，不用每次持有都调用新的，节约内存

// MakeOkReply returns a ok reply
func MakeOkReply() *OkReply {
	return theOkReply
}

// 3 回复错误
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")

func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

//4 回复空数组
var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply is a empty list
type EmptyMultiBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// 5 回复空
type NoReply struct{}

var noBytes = []byte("")

// ToBytes marshal redis.Reply
func (r *NoReply) ToBytes() []byte {
	return noBytes
}
