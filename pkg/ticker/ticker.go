package ticker

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vincoll/vigie/pkg/process"

	log "github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
)

//https://guzalexander.com/2017/05/31/gracefully-exit-server-in-go.html

type TickerPoolManager struct {
	tickerPools map[time.Duration]TickerPool
	ChanToSched chan teststruct.Task
}

func NewTickerPoolManager(toSched chan teststruct.Task) *TickerPoolManager {

	tpm := TickerPoolManager{
		ChanToSched: toSched,
		tickerPools: make(map[time.Duration]TickerPool, 0),
	}
	return &tpm
}

func (tpm *TickerPoolManager) AddTickerPool(freq time.Duration) error {

	if freq <= time.Millisecond {
		return fmt.Errorf("TickerPool cannot be created: frequency cannot be < 1ms")
	}

	tp := TickerPool{
		ticker:          *time.NewTicker(freq),
		frequency:       freq,
		Tasks:           make(map[uint64]*tPoolTasker, 0),
		close:           make(chan struct{}),
		chanToScheduler: tpm.ChanToSched,
	}

	tpm.tickerPools[freq] = tp
	return nil

}

func (tpm *TickerPoolManager) IsTickerPool(freq time.Duration) bool {
	_, present := tpm.tickerPools[freq]
	return present
}

// startEachTickerPool déclenche tout les Tickers afin de débuter les tests.
func (tpm *TickerPoolManager) StartEachTickerPool() {
	for _, tp := range tpm.tickerPools {
		go tp.Start()
	}
}

// stopEachTickerPool stops all the tickers
func (tpm *TickerPoolManager) StopEachTickerPool() {
	// Stop all the tickers
	for _, tp := range tpm.tickerPools {
		tp.Stop()
	}
}

// stopEachTickerPool stops all the tickers
func (tpm *TickerPoolManager) AddTask(t teststruct.Task) {

	tpool := tpm.tickerPools[t.TestStep.ProbeWrap.Frequency]
	tpool.AddTask(t)

}

func (tpm *TickerPoolManager) GracefulShutdown() {
	tpm.StopEachTickerPool()
}

type TickerPool struct {
	ticker          time.Ticker
	frequency       time.Duration
	Tasks           map[uint64]*tPoolTasker
	close           chan struct{}
	chanToScheduler chan teststruct.Task
}

type tPoolTasker struct {
	sync.Mutex
	task            teststruct.Task
	schedulingDelay time.Duration
	// reSync is used when a newState is loaded. => Reset Tickerpools
	// It add a scheduling delay based on the last attempt to respect
	// the test interval of this task even if a tickerpool have been reset.
	reSync time.Duration
}

func (tpt *tPoolTasker) applyReSync() {
	tpt.Lock()
	if tpt.reSync != 0 {
		time.Sleep(tpt.reSync)
		tpt.resetReSync()
	}
	tpt.Unlock()
}

func (tpt *tPoolTasker) resetReSync() {
	tpt.reSync = 0
}

func (tp *TickerPool) AddTask(task teststruct.Task) {

	task3 := tPoolTasker{
		task:            task,
		schedulingDelay: 0,
		reSync:          task.TestStep.GetReSyncro(),
	}

	tp.Tasks[task.TestStep.ID] = &task3

}

// Start the tickerpool as seperate goroutine
func (tp *TickerPool) Start() {

	// Some types of probes will be more likely to have long intervals
	// Avoid waiting the first tick in case of a long duration.

	if tp.frequency > time.Duration(59)*time.Second {

		// Delay the launch of the different tickerpools
		// to avoid a startup spike.

		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(15) // n will be between 0 and val
		wait := time.Duration(n) * time.Second
		time.Sleep(wait)

		utils.Log.WithFields(log.Fields{
			"package": "ticker",
		}).Debugf("Ticker %s PRE-START (%s) at %s", tp.frequency.String(), wait, time.Now().Format(time.RFC3339))

		tp.processAllTasks2()
	}

	go tp.run()

	return
}

// run will run all the tasks on each ticking
func (tp *TickerPool) run() {
	// Loop, Start at each Ticker Tick

	for {
		select {
		case <-tp.ticker.C:
			tp.processAllTasks2()
		case <-tp.close:
			tp.ticker.Stop()
			return
		}
	}

}

// run will run all the tasks on each ticking
func (tp *TickerPool) Stop() {
	// Stop the tick
	tp.close <- struct{}{}
	close(tp.close)

}

// processAllTasks launched at each Tick :
// Run all TickerPool tests in independent goroutines.
// Fine timeout management is managed as close as possible to the probe.
func (tp *TickerPool) processAllTasks2() {

	utils.Log.WithFields(log.Fields{
		"package": "ticker",
	}).Debugf("Ticker %s START at %s with %d Tasks ", tp.frequency.String(), time.Now().Format(time.RFC3339), len(tp.Tasks))

	// Task scheduling is an important part, for now it's very simple and limited.
	// More refined way to avoid spikes will be engaged.

	for i := range tp.Tasks {

		tp.chanToScheduler <- tp.Tasks[i].task

	}
}

// processAllTasks launched at each Tick :
// Run all TickerPool tests in independent goroutines.
// Fine timeout management is managed as close as possible to the probe.
func (tp *TickerPool) processAllTasks() {

	utils.Log.WithFields(log.Fields{
		"package": "ticker",
	}).Debugf("Ticker %s START at %s with %d Tasks ", tp.frequency.String(), time.Now().Format(time.RFC3339), len(tp.Tasks))

	// Task scheduling is an important part, for now it's very simple and limited.
	// More refined way to avoid spikes will be engaged.

	// To avoid spikes at each tick a little pause is interleaved between the tests.
	// The pause scheduling_delay is calculated from 5% freq, otherwise for 24H freq the pause are too long

	// Dummy leapPause calc
	maxInterval := float64(tp.frequency.Nanoseconds()) * 0.05
	leapPause := time.Duration(int64(maxInterval) / int64(len(tp.Tasks)))

	for i := range tp.Tasks {

		go func(t *tPoolTasker) {
			t.applyReSync()
			process.ProcessTask(t.task)
		}(tp.Tasks[i])

		// Sleep leapPause before another run
		time.Sleep(leapPause)
	}
	/*
		for i, _ := range tp.Tasks {

			go func() {
				tp.Tasks[i].applyReSync()
				process.ProcessTask(tp.Tasks[i].task)
			}()

			// Sleep leapPause before another run
			time.Sleep(leapPause)
		}
	*/
}
