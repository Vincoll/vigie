package teststruct

import (
	"fmt"
	"sync"
	"time"
)

type JSONTestCase struct {
	Name      string               `json:"name"`
	Config    configTestStructJson `json:"config"`
	Loop      []string             `json:"loop"`
	JsonSteps []JSONStep           `json:"steps"`
}

// toTestCase is a single test case with its result.
type TestCase struct {
	Mutex       sync.RWMutex
	Config      configTestStruct
	ID          int64
	Name        string
	Status      StepStatus
	CountChange uint
	LastChange  time.Time
	TestSteps   map[int64]*TestStep
}

// importConfig allows configuration inheritance
// Only insert configs not present, in case of different config
// priority is given to the lowest entity.
func importConfig(cfgTS *configTestStruct, cfgTC configTestStruct) configTestStruct {

	for k, v := range cfgTS.Frequency {
		if _, present := cfgTC.Frequency[k]; !present {
			cfgTC.Frequency[k] = v
		}
	}
	for k, v := range cfgTS.Timeout {
		if _, present := cfgTC.Timeout[k]; !present {
			cfgTC.Timeout[k] = v
		}
	}
	for k, v := range cfgTS.Retrydelay {
		if _, present := cfgTC.Retrydelay[k]; !present {
			cfgTC.Retrydelay[k] = v
		}
	}
	if cfgTC.Retry == nil {
		cfgTC.Retry = map[string]int{}
	}
	for k, v := range cfgTS.Retry {
		if _, present := cfgTC.Retry[k]; !present {
			cfgTC.Retry[k] = v
		}
	}

	return cfgTC

}

// https://blog.gopheracademy.com/advent-2016/advanced-encoding-decoding/
func (jtc JSONTestCase) toTestCase(ctsTS *configTestStruct, tsVars map[string][]string) (TestCase, error) {

	var tc TestCase

	if jtc.Name == "" {
		return TestCase{}, fmt.Errorf("testcase name is missing")
	}
	tc.Name = jtc.Name

	if jtc.JsonSteps == nil {
		return TestCase{}, fmt.Errorf("no teststep detected in %q testcase", jtc.Name)
	}

	// UnMarshall spécifique de la config avec conversion de strings type (1d,7m) en time.duration
	ctcTC, err := unmarshallConfigTestStruct(jtc.Config)
	if err != nil {
		return TestCase{}, fmt.Errorf("config declaration: %s", err)
	}

	ctcTC = importConfig(ctsTS, ctcTC)

	// Add TestSteps
	tc.TestSteps = make(map[int64]*TestStep, len(jtc.JsonSteps))

	for _, jStp := range jtc.JsonSteps {

		teststeps, err := jStp.toTestStep(&ctcTC, tsVars)
		if err != nil {
			return TestCase{}, fmt.Errorf("%s : Step is invalid: %s", tc.Name, err)
		}

		tc.addSteps(teststeps)
	}

	return tc, nil
}

// addStep Add a TestStep into this toTestCase
func (tc *TestCase) addSteps(teststeps []TestStep) {
	tc.Mutex.Lock()
	y := int64(len(tc.TestSteps))
	for _, tstep := range teststeps {
		tstp2 := tstep
		tstp2.ID = y
		tc.TestSteps[y] = &tstp2
		y++
	}
	tc.Mutex.Unlock()
}

func (tc *TestCase) FailureCount() (failCount int) {
	tc.Mutex.RLock()
	for _, tstep := range tc.TestSteps {
		tstep.Mutex.RLock()
		failCount += len(tstep.Failures)
		tstep.Mutex.RUnlock()
	}
	tc.Mutex.RUnlock()

	return failCount
}
