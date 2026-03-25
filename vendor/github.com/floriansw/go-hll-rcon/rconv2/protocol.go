package rconv2

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"syscall"
	"time"
)

const (
	magicNumber = uint32(0xDE450508)
)

var (
	ErrWriteSentUnequal    = errors.New("write wrote less or more bytes than command is long")
	ErrReadLengthUnequal   = errors.New("server wrote less bytes than advertised")
	ReconnectTriesExceeded = errors.New("there are no reconnects left")
)

type socket struct {
	con            net.Conn
	pw             string
	host           string
	port           int
	reconnectCount int

	xorKey    []byte
	authToken string

	lastContext *context.Context
	// TODO Completely the wrong place for that, probably. But, the library does not support sending multiple
	// commands over the same socket yet, anyway. Hence it doesn't matter, it just needs to go somewhere for now.
	lastRequestId uint32
}

type Request[T, U any] struct {
	Body T
}

func (r *Request[T, U]) do(s *socket) (result Response[U], err error) {
	b := marshal(r.asRawRequest(s.authToken))
	err = s.write(b)
	if err != nil {
		return result, err
	}
	res, err := s.read()
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(res, &result)
	return result, err
}

func (r *Request[T, U]) asRawRequest(authToken string) rawRequest {
	body := r.Body
	var d []byte
	t := reflect.ValueOf(r.Body)
	if t.Kind() == reflect.String {
		d = []byte(t.String())
	} else {
		d, _ = json.Marshal(body)
	}
	var cmd string
	if c, ok := any(body).(Command); ok {
		cmd = c.CommandName()
	} else {
		cmd = reflect.TypeOf(body).Name()
	}
	return rawRequest{
		Command:   cmd,
		Body:      string(d),
		AuthToken: authToken,
		Version:   2,
	}
}

type Response[T any] struct {
	StatusCode    int    `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
	Version       int    `json:"version"`
	Command       string `json:"name"`
	Content       string `json:"contentBody"`
}

func (r *Response[T]) Body() (res T) {
	if _, ok := any(res).(string); ok {
		return any(r.Content).(T)
	}
	_ = json.Unmarshal([]byte(r.Content), &res)
	return
}

type rawRequest struct {
	Command   string      `json:"Name"`
	AuthToken string      `json:"AuthToken"`
	Body      interface{} `json:"ContentBody"`
	Version   int         `json:"Version"`
}

func (r *socket) SetContext(ctx context.Context) error {
	r.lastContext = &ctx
	if r.con == nil {
		return nil
	}
	if deadline, ok := ctx.Deadline(); ok {
		return r.con.SetDeadline(deadline)
	} else {
		return r.con.SetDeadline(time.Now().Add(20 * time.Second))
	}
}

func (r *socket) Context() context.Context {
	if r.lastContext != nil {
		return *r.lastContext
	}
	return context.Background()
}

func makeConnectionV2(h string, p int) (net.Conn, error) {
	return net.DialTimeout("tcp4", fmt.Sprintf("%s:%d", h, p), 5*time.Second)
}

func newSocket(ctx context.Context, h string, p int, pw string) (*socket, error) {
	r := &socket{
		pw:             pw,
		host:           h,
		port:           p,
		reconnectCount: 0,
	}
	_ = r.SetContext(ctx)
	return r, r.reconnect(nil)
}

func (r *socket) Close() error {
	return r.con.Close()
}

func (r *socket) login() error {
	req := rawRequest{
		Command: "Login",
		Version: 2,
		Body:    r.pw,
	}
	err := r.write(marshal(req))
	if err != nil {
		return err
	}
	res, err := r.read()
	if err != nil {
		return err
	}
	var data Response[string]
	err = json.Unmarshal(res, &data)
	if err != nil {
		return err
	}
	if data.StatusCode == 401 {
		return ErrInvalidCredentials
	} else if data.StatusCode != 200 {
		return NewUnexpectedStatus(data.StatusCode, data.StatusMessage)
	}
	r.authToken = data.Content
	return nil
}

func (r *socket) greatServer() error {
	req := rawRequest{
		Command: "ServerConnect",
		Version: 2,
		Body:    nil,
	}
	err := r.write(marshal(req))
	if err != nil {
		return err
	}
	res, err := r.read()
	if err != nil {
		return err
	}
	var data Response[[]byte]
	err = json.Unmarshal(res, &data)
	if err != nil {
		return err
	}
	if data.StatusCode != 200 {
		return NewUnexpectedStatus(data.StatusCode, data.StatusMessage)
	}
	r.xorKey, err = base64.StdEncoding.AppendDecode(r.xorKey, []byte(data.Content))
	return err
}

func marshal(v rawRequest) []byte {
	req, _ := json.Marshal(v)
	return req
}

func (r *socket) retryableWrite(err error, cmd []byte) error {
	err = r.reconnect(err)
	if err != nil {
		return err
	}
	return r.write(cmd)
}

func (r *socket) write(cmd []byte) error {
	data := r.xor(cmd)
	err := binary.Write(r.con, binary.LittleEndian, []uint32{magicNumber, r.lastRequestId, uint32(len(data))})
	if errors.Is(err, syscall.EPIPE) {
		return r.retryableWrite(err, cmd)
	} else if err != nil {
		return err
	}
	r.lastRequestId++
	s, err := r.con.Write(data)
	if errors.Is(err, syscall.EPIPE) {
		return r.retryableWrite(err, cmd)
	} else if err != nil {
		return err
	}
	if s != len(cmd) {
		return fmt.Errorf("%w Cmd: %s (%d), sent: %d", ErrWriteSentUnequal, cmd, len(cmd), s)
	}

	r.resetReconnectCount()
	return err
}

func (r *socket) reconnect(orig error) error {
	if r.reconnectCount > 3 {
		return ReconnectTriesExceeded
	}
	r.reconnectCount++
	con, err := makeConnectionV2(r.host, r.port)
	if err != nil {
		return err
	}
	r.con = con
	err = r.SetContext(r.Context())
	if err != nil {
		return err
	}
	err = r.greatServer()
	if err != nil {
		return fmt.Errorf("great failed: %s, original error: %w", err.Error(), orig)
	}
	err = r.login()
	if err != nil {
		return fmt.Errorf("login failed: %s, original error: %w", err.Error(), orig)
	}
	return nil
}

func (r *socket) read() ([]byte, error) {
	// each response has a fixed 12-byte header, the first 4 bytes is a magic number, followed by the response Id
	// assigned by the server and the next 4 bytes is the content length of the response body
	// byte format as used in python is: <III
	var magic, responseId, contentLength uint32
	err := binary.Read(r.con, binary.LittleEndian, &magic)
	if err != nil {
		return nil, fmt.Errorf("read magic number failed: %w", err)
	}
	if magic != magicNumber {
		return nil, fmt.Errorf("magic number does not match: expected %d to match %d", magic, magicNumber)
	}
	err = binary.Read(r.con, binary.LittleEndian, &responseId)
	if err != nil {
		return nil, fmt.Errorf("read responseId failed: %w", err)
	}
	err = binary.Read(r.con, binary.LittleEndian, &contentLength)
	if err != nil {
		return nil, fmt.Errorf("read content length failed: %w", err)
	}

	answer := make([]byte, contentLength)
	l, err := io.ReadFull(r.con, answer)
	if len(answer) != l {
		return nil, fmt.Errorf("%w responseId: %d, contentLength: %d, read: %d", ErrReadLengthUnequal, responseId, contentLength, l)
	}

	return r.xor(answer), err
}

func (r *socket) xor(src []byte) []byte {
	if r.xorKey == nil {
		return src
	}

	msg := make([]byte, len(src))
	for i, b := range src {
		msg[i] = b ^ r.xorKey[i%len(r.xorKey)]
	}
	return msg
}

func (r *socket) resetReconnectCount() {
	r.reconnectCount = 0
	r.lastRequestId = 0
}
