package scheduler

import (
	"github.com/vincoll/vigie/pkg/process"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/tsdb"
	"sync"
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

func (s *Scheduler) Submit(t teststruct.Task) {

	s.chProcess <- t

}

func (s *Scheduler) start() {
	go func() {
		for {
			task := <-s.chProcess
			testResult := process.ProcessTask(task)

			// Insert Task ResultStatus to DB
			tsdb.TsdbMgr.WriteOnTsdbs(task, testResult)

		}
	}()
}
