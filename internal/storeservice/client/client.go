package client

import (
	"context"
	"errors"
	"kvstore/internal/common/grpcclient"
	"kvstore/internal/protobuf/storepb"
)

type Client struct {
	conn *grpcclient.GRPCClient
}

func New(cl *grpcclient.GRPCClient) *Client {
	return &Client{
		conn: cl,
	}
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	sc := storepb.NewStoreClient(c.conn.ClientConn)
	resp, err := sc.Get(ctx, &storepb.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}

	return resp.Value, nil
}

func (c *Client) Put(ctx context.Context, key string, value []byte) error {
	sc := storepb.NewStoreClient(c.conn.ClientConn)
	resp, err := sc.Put(ctx, &storepb.PutRequest{Key: key, Value: value})
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return errors.New(resp.Error.Message)
	}

	return nil
}
