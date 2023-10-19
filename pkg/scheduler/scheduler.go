package scheduler

import (
	"github.com/vincoll/vigie/pkg/process"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/tsdb"
	"sync"
	"time"
)

type Scheduler struct {
	mu        sync.RWMutex
	workers   uint64
	chProcess chan teststruct.Task
}

func NewScheduler(chProcess chan teststruct.Task, maxworker uint64) *Scheduler {

	sched := Scheduler{
		workers:   maxworker,
		chProcess: chProcess,
	}
	sched.start()

	return &sched

}

func (s *Scheduler) start() {
	go func() {
		for {
			// Continuous read (multiple senders from tickers)
			task := <-s.chProcess
			go func() {
				testResult := process.ProcessTask(task)
				// Insert Task ResultStatus to DB
				tsdb.TsdbMgr.WriteOnTsdbs(task, testResult)
			}()
		}
	}()
}

// reSync Task in case of a
func reSyncTask(task teststruct.Task) {

	nextCheck := task.TestStep.LastAttempt.Add(task.TestStep.ProbeWrap.Frequency)
	now := time.Now()

	diff := nextCheck.Sub(now)

	if diff.Milliseconds() > 0 {
		time.Sleep(diff)
	}

}
