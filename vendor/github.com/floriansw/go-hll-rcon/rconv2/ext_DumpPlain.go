package rconv2

import "context"

func DumpPlain[T any](ctx context.Context, c *Connection, r T) (*string, error) {
	return execCommand[T, string](ctx, c.socket, r)
}
