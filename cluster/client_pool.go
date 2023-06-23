package cluster

import (
	"context"
	"errors"
	"go-redis/resp/client"

	pool "github.com/jolestar/go-commons-pool/v2"
)

// 一个连接池，准确说是连接工厂，加工连接池
type connectionFactory struct {
	Peer string //连接的结点地址
}

func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	c, err := client.MakeClient(f.Peer) //传入地址
	if err != nil {
		return nil, err
	}
	c.Start() //开始登录建立连接
	return pool.NewPooledObject(c), nil
}

func (f *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	c, ok := object.Object.(*client.Client)
	if !ok { //连接池类型不匹配
		return errors.New("type mismatch")
	}
	c.Close() //关闭连接
	return nil
}

func (f *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	// do validate
	return true
}

func (f *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do activate
	return nil
}

func (f *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do passivate
	return nil
}
