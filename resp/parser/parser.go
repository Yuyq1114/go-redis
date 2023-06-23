// redis解析器，解析用户发送的请求，如$ * \r\n等
package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// 解析后的数据
type Payload struct {
	Data resp.Reply //用户和服务端发送的数据结构一样，都是reply
	Err  error
}

// 解析器的状态
type readState struct {
	readingMultiLine  bool //正在解析单行还是多行
	expectedArgsCount int  //用户命令参数个数
	msgType           byte
	args              [][]byte //用户传输数据本身
	bulkLen           int64    //记录将传输字节长度
}

// 计算解析有无完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// 异步解析，让核心业务处理和解析并发执行
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// 解析核心方法
func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() { //做recover，不要让他在异常情况下退出
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state) //调用一，读一行数据
		if err != nil {
			if ioErr { //如果是io错误
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			ch <- &Payload{ //如果是协议错误
				Err: err,
			}
			state = readState{} //状态置空，继续读下一行
			continue
		}
		//根据该结构体内容判断是否是多行解析模式
		if !state.readingMultiLine {
			if msg[0] == '*' { //如果用户发送第一行是*
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error:" + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}

			} else if msg[0] == '$' { //如果一开始遇见的是$
				err := parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error:" + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 { //-1说明用户是空指令
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}

			} else {
				result, err := parsrseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else { // 是多行模式
			err := readBody(msg, &state)
			if err != nil {
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error:" + string(msg)),
					}
					state = readState{}
					continue
				}
			}
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}

	}
}

// 解析辅助函数一，在io中取出一行
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {

	var msg []byte
	var err error
	if state.bulkLen == 0 { //1.无$, 按\r\n切分
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error:" + string(msg))
		}

	} else { //2.存在$,严格读取字符个数
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil

}

// 解析辅助函数二，如果readLine读到*，但是不知道具体意思,需要本函数改变状态
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32) //输入字符，把字符代表的数字取出
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 { //如果输入大小是0
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 { //如果大于0，正常解析
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else { //如果小于0，错误
		return errors.New("protocol error: " + string(msg))
	}

}

// 解析辅助函数三，如果用户输入是以$开头的单行字符串，修改解析器状态

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// 解析辅助函数四，如客户端发送+OK，-err等
func parsrseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n") //修剪\r\n后缀
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)

	}
	return result, nil

}

// 解析辅助函数五，解析发送内容体
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	// 处理的$3类似的
	if line[0] == '$' {
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else { //处理的是PING
		state.args = append(state.args, line)
	}
	return nil
}
