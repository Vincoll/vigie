package worker

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"sync"

	"github.com/vincoll/vigie/internal/worker/utils"
)

type Worker struct {
	mu       sync.RWMutex
	Status   string // NotReady, Ready, WaitElection, Healthy => TODO const enum
	HostInfo HostInfo
	CacheDNS *utils.ResolverCache
}

// NewWorker Constructor of Vigie
func NewWorker() *Worker {

	w := Worker{
		Status: "NotReady",
	}

	return &w
}

func (w *Worker) Start() error {

	// Create folder structure
	err := w.createTempFolder()
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) gracefulShutdown() {

}

func (w *Worker) Health() (status string) {
	w.mu.Lock()
	status = w.Status
	w.mu.Unlock()
	return status
}

// createTempFolder Create a temp folder required for some probes
func (w *Worker) createTempFolder() error {

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

		path = fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\Temp\\webapi\\", username)

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
