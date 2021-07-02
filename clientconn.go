package grpc4go

import (
	"context"
	"errors"
	"github.com/smartwalle/pool4go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"time"
)

var (
	ErrServerNotFound = errors.New("server not found")
)

type ClientConn struct {
	pool       pool4go.Pool
	maxRetries int
}

func Dial(target string, poolSize int, timeout time.Duration, opts ...grpc.DialOption) *ClientConn {
	if poolSize <= 0 {
		poolSize = 1
	}
	var c = &ClientConn{}
	c.pool = pool4go.New(func() (pool4go.Conn, error) {
		var ctx = context.Background()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()

			opts = append(opts, grpc.WithBlock())
		}
		return grpc.DialContext(ctx, target, opts...)
	}, pool4go.WithMaxIdle(poolSize), pool4go.WithMaxOpen(poolSize))
	c.maxRetries = poolSize
	return c
}

func (this *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	for i := 0; i <= this.maxRetries; i++ {
		var conn, err = this.pool.Get()
		if err != nil {
			return err
		}

		var nConn = conn.(*grpc.ClientConn)

		var state = nConn.GetState()
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			this.pool.Release(conn)
			continue
		}

		err = nConn.Invoke(ctx, method, args, reply, opts...)
		this.pool.Put(conn)
		return err
	}
	return ErrServerNotFound
}

func (this *ClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	for i := 0; i <= this.maxRetries; i++ {
		var conn, err = this.pool.Get()
		if err != nil {
			return nil, err
		}

		var nConn = conn.(*grpc.ClientConn)

		var state = nConn.GetState()
		if state == connectivity.TransientFailure || state == connectivity.Shutdown {
			this.pool.Release(conn)
			continue
		}

		stream, err := nConn.NewStream(ctx, desc, method, opts...)
		this.pool.Put(conn)
		return stream, err
	}
	return nil, ErrServerNotFound
}

func (this *ClientConn) Close() {
	this.pool.Close()
}
