package ha

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/vincoll/vigie/pkg/utils"
)

func (cc *ConsulClient) GetHealthChecks(state string, options *consul.QueryOptions) ([]*consul.HealthCheck, error) {
	checks, _, err := cc.Consul.Health().State("any", options)
	return checks, err
}

func (cc *ConsulClient) GetSession(sessionName string) string {
	name := cc.GetAgentName()
	sessions, _, err := cc.Consul.Session().List(nil)
	for _, session := range sessions {
		if session.Name == sessionName && session.Node == name {
			return session.ID
		}
	}

	utils.Log.Infof("No leadership sessions found, creating...")

	sessionEntry := &consul.SessionEntry{Name: sessionName}
	session, _, err := cc.Consul.Session().Create(sessionEntry, nil)
	if err != nil {
		utils.Log.Warn(err)
	}
	return session
}

func (cc *ConsulClient) AquireSessionKey(key string, session string) (bool, error) {

	pair := &consul.KVPair{
		Key:     key,
		Value:   []byte(cc.GetAgentName()),
		Session: session,
	}

	aquired, _, err := cc.Consul.KV().Acquire(pair, nil)

	return aquired, err
}

func (cc *ConsulClient) GetAgentName() string {
	agent, _ := cc.Consul.Agent().Self()
	return agent["Config"]["NodeName"].(string)
}

func (cc *ConsulClient) PutKey(key *consul.KVPair) error {
	_, err := cc.Consul.KV().Put(key, nil)
	return err
}

func (cc *ConsulClient) GetKey(keyName string) (*consul.KVPair, error) {
	kv, _, err := cc.Consul.KV().Get(keyName, nil)
	return kv, err
}

func (cc *ConsulClient) ReleaseKey(key *consul.KVPair) (bool, error) {
	released, _, err := cc.Consul.KV().Release(key, nil)
	return released, err
}
