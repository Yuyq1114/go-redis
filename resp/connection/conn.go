package connection

import (
	"go-redis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// 维护用户连接
type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait
	mu           sync.Mutex
	selectedDB   int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn, //socket连接
	}
}

// 远程连接地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// 关闭客户端
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second) //超时关闭
	_ = c.conn.Close()
	return nil
}

// 把字节组从connetcion写回客户端
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// 返回当前用户DB
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// 修改当前DB
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
