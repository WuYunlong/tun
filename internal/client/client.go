package client

import "context"

type Client struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
}

func NewClient() *Client {
	tc := new(Client)
	return tc
}

func (tc *Client) Run(ctx context.Context) {
	tc.ctx, tc.cancel = context.WithCancelCause(ctx)

	<-tc.ctx.Done()
	tc.stop()
}

func (tc *Client) stop() {
}
