package ha

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/utils"
	"os"
	"sync"
	"time"
)

// Global Var for now, to avoid time-consuming modifications
// in case of a change of code rearchitecture.
// Consul state will be injected later

// https://pkg.go.dev/github.com/hashicorp/consul/api?tab=doc

type ConsulClient struct {
	mu       sync.RWMutex
	Consul   *consul.Client
	Election *Election
}

type ConfConsul struct {
	Enable bool `toml:"enable"`
	// Address is the address of the Consul server
	Address string `toml:"Address"`

	// Scheme is the URI scheme for the Consul server
	Scheme string

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string `toml:"token"`
}

type notify struct {
	T string
}

func (n *notify) EventLeader(e bool) {
	if e {
		utils.Log.WithFields(log.Fields{
			"package": "ha",
		}).Infof("This %s instance is now a leader", n.T)
	} else {
		utils.Log.WithFields(log.Fields{
			"package": "ha",
		}).Infof("This %s instance is no longer a leader", n.T)
	}
}

func InitHAConsul(confConsul ConfConsul, apiPort int) (*ConsulClient, error) {

	cfg := &consul.Config{
		Address:   confConsul.Address,
		Scheme:    confConsul.Scheme,
		Token:     confConsul.Token,
		TLSConfig: consul.TLSConfig{},
	}

	// Get a new client
	client, err := consul.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	n := &notify{
		T: "Vigie",
	}

	cc := ConsulClient{Consul: client}
	err = cc.DeRegister("vigie")
	err = cc.register(apiPort)
	if err != nil {
		return nil, err
	}

	elconf := &ElectionConfig{
		CheckTimeout: 5 * time.Second,
		Client:       cc.Consul,
		Checks:       []string{"service:health"},
		Key:          "service/election/leader",
		LogLevel:     LogDebug,
		Event:        n,
	}
	e := NewElection(elconf)

	// Start election in background, this can take ~5sec
	// The result of the election will be fetch later
	go e.Init()

	return &cc, nil
}

func (cc *ConsulClient) GracefulShutdown() {

}

func (cc *ConsulClient) isLeader() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.Election.IsLeader()
}

func (cc *ConsulClient) test() {

	// Get a handle to the KV API
	kv := cc.Consul.KV()

	x, _, err2 := kv.List("vigie", nil)
	if err2 != nil {
		panic(err2)
	}
	fmt.Sprint(len(x))

	// PUT a new KV pair
	p := &consul.KVPair{Key: "12/REDIS_MAXCLIENTS", Value: []byte("1000")}
	_, err := kv.Put(p, nil)
	if err != nil {
		panic(err)
	}

	// Lookup the pair
	pair, _, err := kv.Get("REDIS_MAXCLIENTS", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("KV: %v %s\n", pair.Key, pair.Value)

}

// DeRegister a service with consul local agent
func (cc *ConsulClient) DeRegister(id string) error {
	return cc.Consul.Agent().ServiceDeregister(id)
}

func (cc *ConsulClient) register(apiPort int) error {

	host, err := os.Hostname()
	if err != nil {
		return err
	}

	host = "vigiehost"

	check := &consul.AgentServiceCheck{
		CheckID:                        "",
		Name:                           "",
		Args:                           nil,
		DockerContainerID:              "",
		Shell:                          "",
		Interval:                       "5s",
		Timeout:                        "3s",
		TTL:                            "",
		HTTP:                           fmt.Sprintf("http://%s:%v/metrics", host, apiPort),
		Header:                         nil,
		Method:                         "",
		Body:                           "",
		TCP:                            "",
		Status:                         "",
		Notes:                          "",
		TLSSkipVerify:                  false,
		GRPC:                           "",
		GRPCUseTLS:                     false,
		AliasNode:                      "",
		AliasService:                   "",
		DeregisterCriticalServiceAfter: "",
	}

	reg := &consul.AgentServiceRegistration{
		ID:      "health",
		Name:    "vigie",
		Port:    apiPort,
		Address: host,
		Check:   check,
	}
	_ = cc.Consul.Agent().ServiceRegister(reg)

	return nil

}

/// ---
