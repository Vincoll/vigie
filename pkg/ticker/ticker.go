package ticker

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/process"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
)

type TickerPool struct {
	ticker    time.Ticker
	frequency time.Duration
	tasks     []teststruct.Task
}

func NewTickerPool(freq time.Duration) (*TickerPool, error) {

	var tp TickerPool

	if freq <= time.Millisecond {
		return nil, fmt.Errorf("TickerPool cannot be created: frequency cannot be < 1ms")
	} else {
		tp.frequency = freq
		tp.ticker = *time.NewTicker(freq)
		tp.tasks = make([]teststruct.Task, 0)
	}
	return &tp, nil
}

func (tp *TickerPool) AddTask(tsuite *teststruct.TestSuite, tcase *teststruct.TestCase, tstep *teststruct.TestStep) {

	ntask := teststruct.Task{
		TestSuite: tsuite,
		TestCase:  tcase,
		TestStep:  tstep,
	}

	tp.tasks = append(tp.tasks, ntask)
}

// Start the tickerpool as seperate goroutine
func (tp *TickerPool) Start() {

	// Some types of probes will be more likely to have long intervals
	// Avoid waiting the first tick in case of a long duration.

	if tp.frequency > time.Duration(59)*time.Second {

		// Delay the launch of the different tickerpools
		// to avoid a startup spike.

		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(10) // n will be between 0 and 10
		wait := time.Duration(n) * time.Second
		time.Sleep(wait)

		utils.Log.WithFields(log.Fields{
			"package": "ticker",
		}).Debugf("Ticker %s PRE-START (%s) at %s", tp.frequency.String(), wait, time.Now().Format(time.RFC3339))

		tp.processAllTasks()
	}
	go tp.run()
}

// run will run all the tasks on each ticking
func (tp *TickerPool) run() {
	// Should not be Stop: Infinite Loop, Start at each Ticker Tick
	for {
		select {
		case <-tp.ticker.C:
			tp.processAllTasks()
		}
	}
}

// processAllTasks lancé à chaque Tick :
// Lance tout les tests de la TickerPool dans des goroutines indépendantes.
// Normalement la gestion fine du timeout est géré au plus proche de la probe
func (tp *TickerPool) processAllTasks() {

	utils.Log.WithFields(log.Fields{
		"package": "ticker",
	}).Debugf("Ticker %s START at %s", tp.frequency.String(), time.Now().Format(time.RFC3339))

	// Dummy interval calc
	interval := time.Duration(tp.frequency.Nanoseconds() / int64(len(tp.tasks)))

	for i := range tp.tasks {
		go func(t teststruct.Task) { process.ProcessTask(t) }(tp.tasks[i])
		// Sleep interval before another run
		time.Sleep(interval)
	}

}
