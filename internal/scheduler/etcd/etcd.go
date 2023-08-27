package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ConfETCD struct {
	Enable bool   `toml:"enable"`
	URL    string `toml:"url"`
}

type ETCDClient struct {
	conf         ConfETCD
	client       *clientv3.Client
	IngoingTests chan string
	status       string
	logger       *zap.SugaredLogger
}

func NewClient(ctx context.Context, conf ConfETCD, logger *zap.SugaredLogger) (*ETCDClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	ec := ETCDClient{client: cli, status: "running"}

	return &ec, nil
}

func (e ETCDClient) GracefulShutdown() error {

	e.logger.Infow("ETCD Client is gracefully shutting down")
	err := e.client.Close()
	if err != nil {
		return err
	}
	return nil
}
