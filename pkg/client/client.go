// Package client is the Go client other services use to send notifications through
// comms-proxy's Notifier API (cluster-internal gRPC).
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gofreego/comms-proxy/api/commsproxy_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	// Host is comms-proxy's gRPC address, e.g. comms-proxy.zshala-com.svc.cluster.local:8081.
	Host string `yaml:"Host"`
	// Timeout bounds each Send; 0 means 5s.
	Timeout time.Duration `yaml:"Timeout"`
}

type Client struct {
	conn     *grpc.ClientConn
	notifier commsproxy_v1.NotifierClient
	timeout  time.Duration
}

// New connects lazily: no call is made until the first Send.
func New(cfg *Config) (*Client, error) {
	conn, err := grpc.NewClient(cfg.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("comms-proxy client: %w", err)
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{conn: conn, notifier: commsproxy_v1.NewNotifierClient(conn), timeout: timeout}, nil
}

// Send delivers one notification; see commsproxy_v1.SendRequest.
func (c *Client) Send(ctx context.Context, req *commsproxy_v1.SendRequest) (*commsproxy_v1.SendResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.notifier.Send(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
