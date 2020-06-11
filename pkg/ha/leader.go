package ha

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/vincoll/vigie/pkg/utils"
	"log"
	"runtime"
	"sync"
	"time"
)

// Code from https://github.com/dmitriyGarden/consul-leader-election

// Log levels
const (
	LogDisable = iota
	LogError
	LogInfo
	LogDebug
)

// Election implements to detect a leader in a cluster of services
type Election struct {
	Client       *api.Client // Consul client
	Checks       []string    // Slice of associated health checks
	leader       bool        // Flag of a leader
	Kv           string      // Key in Consul kv
	sessionID    string      // Id of session
	logLevel     uint8       //  Log level LogDisable|LogError|LogInfo|LogDebug
	inited       bool        // Flag of init.
	CheckTimeout time.Duration
	LogPrefix    string        // Prefix for a log
	stop         chan struct{} // chnnel to stop process
	success      chan struct{} // channel for the signal that the process is stopped
	Event        Notifier
	sync.RWMutex
}

// Notifier can tell your code the event of the leader's status change
type Notifier interface {
	EventLeader(e bool) // The method will be called when the leader status is changed
}

// ElectionConfig config for Election
type ElectionConfig struct {
	Client       *api.Client // Consul client
	Checks       []string    // Slice of associated health checks
	Key          string      // Key in Consul KV
	LogLevel     uint8       // Log level LogDisable|LogError|LogInfo|LogDebug
	LogPrefix    string      // Prefix for a log
	Event        Notifier
	CheckTimeout time.Duration
}

// IsLeader check a leader
func (e *Election) IsLeader() bool {
	e.RLock()
	defer e.RUnlock()
	return e.leader
}

// SetLogLevel is setting level according constants LogDisable|LogError|LogInfo|LogDebug
func (e *Election) SetLogLevel(level uint8) {
	e.logLevel = level
}

// Params: Consul client, slice of associated health checks, service name
func NewElection(c *ElectionConfig) *Election {
	e := &Election{
		Client:       c.Client,
		Checks:       append(c.Checks, "serfHealth"),
		leader:       false,
		Kv:           c.Key,
		CheckTimeout: c.CheckTimeout,
		LogPrefix:    c.LogPrefix,
		stop:         make(chan struct{}),
		success:      make(chan struct{}),
		Event:        c.Event,
	}
	return e
}

func (e *Election) createSession() (err error) {
	ses := &api.SessionEntry{
		Checks: e.Checks,
		TTL:    (3 * e.CheckTimeout).String(),
	}
	e.sessionID, _, err = e.Client.Session().Create(ses, nil)
	if err != nil {
		utils.Log.Errorf("Create session error " + err.Error())
	}
	return
}

func (e *Election) checkSession() (bool, error) {

	if e.sessionID == "" {
		return false, nil
	}
	res, _, err := e.Client.Session().Info(e.sessionID, nil)

	if err != nil {
		utils.Log.Errorf("Info session error " + err.Error())
	}

	return res != nil, err
}

// Try to acquire
func (e *Election) acquire() (bool, error) {
	kv := &api.KVPair{
		Key:     e.Kv,
		Session: e.sessionID,
		Value:   []byte(e.sessionID),
	}
	res, _, err := e.Client.KV().Acquire(kv, nil)
	if err != nil {
		utils.Log.Errorf("Acquire kv error " + err.Error())
	}
	return res, err
}

func (e *Election) disableLeader() {
	e.Lock()
	if e.leader {
		e.leader = false
		utils.Log.Debugf("I'm not a leader.:(")
		if e.Event != nil {
			e.Event.EventLeader(false)
		}
	}
	e.Unlock()
}

func (e *Election) getKvSession() (string, error) {
	p, _, err := e.Client.KV().Get(e.Kv, nil)
	if err != nil {
		utils.Log.Errorf("Kv error " + err.Error())
		return "", err
	}
	if p == nil {
		return "", nil
	}
	return p.Session, nil
}

// Init starting election process
func (e *Election) Init(wg *sync.WaitGroup) {
	defer wg.Done()
	e.Lock()
	if e.inited {
		e.Unlock()
		utils.Log.Infof("Only one init available")
		return
	}
	e.inited = true
	e.Unlock()
	for {
		if !e.isInit() {
			break
		}
		e.process()
		if !e.isInit() {
			break
		}
		wait(e.CheckTimeout)
	}
	utils.Log.Debugf("I'm finished")
}

// Start re-election
func (e *Election) ReElection() error {
	s, err := e.getKvSession()
	if s != "" {
		err = e.destroySession(s)
		if err != nil {
			return fmt.Errorf("%s, %s", err, err)
		}
	}
	return err
}

func (e *Election) destroySession(sesID string) error {
	_, err := e.Client.Session().Destroy(sesID, nil)
	if err != nil {
		utils.Log.Errorf("Destroy session error " + err.Error())
	}
	return err
}

func (e *Election) destroyCurrentSession() (err error) {
	if e.sessionID != "" {
		err = e.destroySession(e.sessionID)
		e.sessionID = ""
	}
	return
}

func (e *Election) isNeedAquire() bool {
	var res string
	var err error
	for {
		if !e.isInit() {
			break
		}
		res, err = e.getKvSession()
		if err != nil {
			e.disableLeader()
			wait(e.CheckTimeout)
		} else {
			break
		}

	}
	if e.sessionID != "" && e.sessionID == res {
		e.enableLeader()
	}
	if res == "" || res != e.sessionID {
		e.disableLeader()
	}

	return res == ""
}

func (e *Election) process() {
	e.waitSession()
	if !e.leader {
		if !e.isNeedAquire() {
			return
		}
		utils.Log.Debugf("Try to acquire")
		res, err := e.acquire()
		if res && err == nil {
			e.enableLeader()
		}
	}
}

func (e *Election) enableLeader() {
	e.Lock()
	if e.isInit() {
		e.leader = true
		utils.Log.Debugf("I'm now a leader !!!")
		if e.Event != nil {
			e.Event.EventLeader(true)
		}
	}
	e.Unlock()
}

// Stop election process
func (e *Election) Stop() {
	e.RLock()
	if !e.inited {
		e.RUnlock()
		return
	}
	e.RUnlock()
	e.stop <- struct{}{}
	<-e.success
}

func (e *Election) isInit() bool {
	for {
		select {
		case <-e.stop:
			e.inited = false
			utils.Log.Debugf("Stop signal recieved")
			e.disableLeader()
			e.destroyCurrentSession()
			e.success <- struct{}{}
			utils.Log.Debugf("Send success")
		default:
			return e.inited
		}
	}
}

func (e *Election) waitSession() {
	for {
		if !e.isInit() {
			break
		}
		isset, err := e.checkSession()

		if isset {
			e.Client.Session().Renew(e.sessionID, nil)
			break
		}
		e.disableLeader()
		if err != nil {
			utils.Log.Debugf("Try to get session info again.")
			if !e.isInit() {
				break
			}
			wait(e.CheckTimeout)
			continue
		}
		err = e.createSession()

		if err == nil {
			utils.Log.Debugf("Session " + e.sessionID + " created")
			break
		}
		if !e.isInit() {
			break
		}
		wait(e.CheckTimeout)
	}
}

func wait(t time.Duration) {
	runtime.Gosched()
	time.Sleep(t)
}

func (e *Election) logError(err string) {
	if e.logLevel >= LogError {
		log.Println(e.LogPrefix + " [ERROR] " + err)
	}
}

func (e *Election) logDebug(s string) {
	if e.logLevel >= LogDebug {
		log.Println(e.LogPrefix + " [DEBUG] " + s)
	}
}

func (e *Election) logInfo(s string) {
	if e.logLevel >= LogInfo {
		log.Println(e.LogPrefix + " [INFO] " + s)
	}
}
