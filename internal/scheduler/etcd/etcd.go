package etcd

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
)

type ConfETCD struct {
	Enable    bool     `toml:"enable"`
	Endpoints []string `toml:"endpoints"`
}

type ETCDClient struct {
	conf         ConfETCD
	client       *clientv3.Client
	elec         *concurrency.Election
	IngoingTests chan string
	status       string
	IsLeader     bool
	logger       *zap.SugaredLogger
}

func NewClient(ctx context.Context, conf ConfETCD, logger *zap.SugaredLogger) (*ETCDClient, error) {

	conf.Endpoints = []string{"localhost:2379", "localhost:22379", "localhost:32379"}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	ec := ETCDClient{client: cli, status: "running", logger: logger, IsLeader: false}

	return &ec, nil
}

func (e ETCDClient) LeaderElection() error {
	// create a new session for leader election
	electionSession, err := concurrency.NewSession(e.client, concurrency.WithTTL(1))
	if err != nil {
		return fmt.Errorf("cannot create a new session for leader election: %s", err)
	}
	defer electionSession.Close() // cleanup

	e.elec = concurrency.NewElection(electionSession, "/election-prefix")
	ctx := context.Background()

	fmt.Println("Attempting to become a leader")
	// start leader election
	if err := e.elec.Campaign(ctx, "value"); err != nil {
		return fmt.Errorf("leader election session started; but the campaign failed: %s", err)

	}
	e.IsLeader = true
	e.logger.Infof("This instance is a now a leader")

	return nil
}

func (e ETCDClient) GracefulShutdown() error {

	var leaderMsg string
	if e.IsLeader {
		leaderMsg = ", as leader this instance is resigning"
		e.elec.Resign(context.Background())
	}
	e.logger.Infof("ETCD Client is gracefully shutting down%s", leaderMsg)

	err := e.client.Close()
	if err != nil {
		return err
	}
	return nil
}
