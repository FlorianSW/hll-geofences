package rconv2

import (
	"context"
)

// Connection represents a persistent connection to a HLL server using RCon. It can be used to issue commands against
// the HLL server and query data.
//
// A Connection is not thread-safe by default. Do not attempt to run multiple commands in different threads or go-routines.
// Doing so may either run into non-expected indefinitely blocking execution (until the context.Context
// deadline exceeds) or to mixed up data (sending a command and getting back the response for another command).
// Instead, in goroutines, use a ConnectionPool and request a new connection for each goroutine. The ConnectionPool will
// ensure that one Connection is only used once at the same time. It also speeds up processing by opening a number of
// Connections until the pool size is reached.
type Connection struct {
	id     string
	socket *socket
}

func execCommand[T, U any](ctx context.Context, so *socket, req T) (result *U, err error) {
	err = so.SetContext(ctx)
	if err != nil {
		return nil, err
	}
	r := Request[T, U]{
		Body: req,
	}
	res, err := r.do(so)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, NewUnexpectedStatus(res.StatusCode, res.StatusMessage)
	}
	return new(res.Body()), nil
}
