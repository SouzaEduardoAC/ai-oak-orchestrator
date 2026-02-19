package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type Transport interface {
	Send(msg interface{}) error
	Listen(handler func(json.RawMessage))
}

type Client struct {
	transport Transport
	pending   map[string]chan *Response
	mu        sync.Mutex
	nextID    int
}

func NewClient(t Transport) *Client {
	c := &Client{
		transport: t,
		pending:   make(map[string]chan *Response),
	}
	go t.Listen(c.handleIncoming)
	return c
}

func (c *Client) handleIncoming(raw json.RawMessage) {
	var resp Response
	if err := json.Unmarshal(raw, &resp); err == nil && resp.ID != nil {
		idStr := string(*resp.ID)
		c.mu.Lock()
		ch, ok := c.pending[idStr]
		if ok {
			delete(c.pending, idStr)
			ch <- &resp
		}
		c.mu.Unlock()
		return
	}
	// TODO: Handle Notifications (Method without ID)
}

func (c *Client) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	idRaw := json.RawMessage(fmt.Sprintf("%d", id))
	idStr := string(idRaw)

	ch := make(chan *Response, 1)
	c.pending[idStr] = ch
	c.mu.Unlock()

	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		ID:      &idRaw,
	}

	if params != nil {
		p, err := json.Marshal(params)
		if err != nil {
			return err
		}
		req.Params = p
	}

	if err := c.transport.Send(req); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, idStr)
		c.mu.Unlock()
		return ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return fmt.Errorf("rpc error (%d): %s", resp.Error.Code, resp.Error.Message)
		}
		if result != nil {
			return json.Unmarshal(resp.Result, result)
		}
		return nil
	}
}
