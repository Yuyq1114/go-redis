package reply

//1 错误回复
type UnknownErrReply struct {
}

var unknownErrBytes = []byte("-Err unknown\r\n")

// 返回给客户的字节
func (r *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

//服务端报错
func (r *UnknownErrReply) Error() string {
	return "Err unknown"
}

//2 cmd命令错误回复
type ArgNumErrReply struct {
	Cmd string
}

// ToBytes marshals redis.Reply
func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

// MakeArgNumErrReply represents wrong number of arguments for command
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

//3 语法错误

type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

// MakeSyntaxErrReply creates syntax error
func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// ToBytes marshals redis.Reply
func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func (r *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

// 4 数据类型错误

type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

// ToBytes marshals redis.Reply
func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

//接口协议错误

type ProtocolErrReply struct {
	Msg string
}

// ToBytes marshals redis.Reply
func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + r.Msg
}
