package grpc4go

import (
	"context"
	"github.com/smartwalle/pool4go"
	"google.golang.org/grpc"
)

type ClientConn struct {
	target string
	opts   []grpc.DialOption
	pool   pool4go.Pool
}

func NewClientConn(target string, poolSize int, opts ...grpc.DialOption) *ClientConn {
	if poolSize <= 0 {
		poolSize = 1
	}
	var c = &ClientConn{}
	c.target = target
	c.opts = opts
	c.pool = pool4go.New(func() (pool4go.Conn, error) {
		return grpc.Dial(target, opts...)
	}, pool4go.WithMaxIdle(poolSize), pool4go.WithMaxOpen(poolSize))
	return c
}

func (this *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	var conn, err = this.pool.Get()
	if err != nil {
		return err
	}

	var nConn = conn.(*grpc.ClientConn)
	err = nConn.Invoke(ctx, method, args, reply, opts...)
	this.pool.Put(conn)
	return err
}

// NewStream 目前还不能直接使用
func (this *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var conn, err = this.pool.Get()
	if err != nil {
		return nil, err
	}

	var nConn = conn.(*grpc.ClientConn)
	stream, err := nConn.NewStream(ctx, desc, method, opts...)
	this.pool.Put(conn)
	return stream, err
}

func (this *ClientConn) Close() {
	this.pool.Close()
}
