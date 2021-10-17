package vigie

import (
	"fmt"
	"github.com/vincoll/vigie/pkg/ha"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/scheduler"
	"github.com/vincoll/vigie/pkg/tsdb"
	"github.com/vincoll/vigie/pkg/utils"

	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/ticker"
	"sync"
)

type Vigie struct {
	mu                sync.RWMutex
	TestSuites        map[uint64]*teststruct.TestSuite
	Status            string // NotReady, Ready, WaitElection, Healthy => TODO const enum
	HostInfo          HostInfo
	ImportManager     load.ImportManager
	ConsulClient      *ha.ConsulClient
	Scheduler         *scheduler.Scheduler
	TsdbManager       *tsdb.Manager
	TickerPoolManager *ticker.TickerPoolManager
	incomingTests     chan map[uint64]*teststruct.TestSuite
}

// NewVigie Constructor of Vigie
func NewVigie() (*Vigie, error) {

	// Chans
	chanToScheduler := make(chan teststruct.Task)
	// Insert Chan Before (PoC) for now
	chanImportMgr := make(chan map[uint64]*teststruct.TestSuite)

	v := &Vigie{
		TestSuites:        map[uint64]*teststruct.TestSuite{},
		TickerPoolManager: ticker.NewTickerPoolManager(chanToScheduler),
		Scheduler:         scheduler.NewScheduler(chanToScheduler, 999),
		incomingTests:     chanImportMgr,
		Status:            "NotReady",
	}

	// Create folder structure
	err := v.createTempFolder()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Vigie) GracefulShutdown() {

	v.ImportManager.GracefulShutdown()
	v.ConsulClient.GracefulShutdown()

}

func (v *Vigie) Health() (status string) {
	v.mu.Lock()
	status = v.Status
	v.mu.Unlock()
	return status
}

// createTempFolder Create a temp folder required for some probes
func (v *Vigie) createTempFolder() error {

	var path string

	if runtime.GOOS == "windows" {

		// prepare Vigie Folder
		// Get User Info
		usr, err := user.Current()
		if err != nil {
			return fmt.Errorf("cannot create a temp folder for Vigie : %s", err.Error())
		}

		tokuser := strings.Split(usr.Username, "\\")
		username := tokuser[len(tokuser)-1]

		path = fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\Temp\\vigie\\", username)

	} else if runtime.GOOS == "linux" {
		path = "/tmp/vigie_dl/"
	}

	// Create Vigie Folder
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0744)
		if err != nil {
			return fmt.Errorf("cannot create a temp folder for Vigie : %s", err.Error())
		}
	}
	// Set this path as global State Var
	utils.TEMPPATH = path

	return nil
}
