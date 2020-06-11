package vigie

import (
	"fmt"
	"github.com/vincoll/vigie/pkg/ha"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/utils"

	"os"
	"os/user"
	"runtime"
	"strings"

	"sync"
	"time"

	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/ticker"
)

type Vigie struct {
	mu            sync.RWMutex
	TestSuites    map[uint64]*teststruct.TestSuite
	tickerpools   map[time.Duration]*ticker.TickerPool
	Status        string // NotReady, Ready, WaitElection, Healthy => TODO const enum
	HostInfo      HostInfo
	ImportManager load.ImportManager
	ConsulClient  *ha.ConsulClient
}

// NewVigie Constructor: Vigie
func NewVigie() (*Vigie, error) {
	v := &Vigie{
		tickerpools: map[time.Duration]*ticker.TickerPool{},
		Status:      "NotReady",
	}
	// Init
	v.TestSuites = map[uint64]*teststruct.TestSuite{}
	v.tickerpools = map[time.Duration]*ticker.TickerPool{}

	// Create folder structure
	err := v.createTempFolder()
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (v *Vigie) Health() (status string) {
	v.mu.Lock()
	status = v.Status
	v.mu.Unlock()
	return status
}

// Add a new TickerPool
func (v *Vigie) createTickerPool(freq time.Duration) error {

	tp, err := ticker.NewTickerPool(freq)
	if err != nil {
		return fmt.Errorf("can not create a Tickerpool: %s", err.Error())
	}

	v.tickerpools[freq] = tp
	return nil
}

// Is TickerPool Exist
func (v *Vigie) getTickerPool(frequency time.Duration) bool {
	_, present := v.tickerpools[frequency]
	return present
}

// Create a temp folder required for some probes
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
