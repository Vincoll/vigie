package load

import (
	"encoding/json"
	"strconv"
	"strings"
)

// getConsulTestScheduling returns the scheduling test decision of the Leader
// regarding the actual followers/workers
func (im *ImportManager) getConsulTestScheduling() (*testSchedulerJSON, error) {

	kv, err := im.ConsulClient.GetKey(consulTestScheduling)
	if err != nil {
		return nil, err
	}

	var tsch testSchedulerJSON

	if err := json.Unmarshal(kv.Value, &tsch); err != nil {
		panic(err)
	}
	return &tsch, nil
}

func (im *ImportManager) pullTstepsFromConsul(tsch *testSchedulerJSON) error {

	/*
		machineID := im.ConsulClient.GetAgentName()

		// Get values of this only Vigie instance
		testsIDs := tsch.value[machineID]

		testSuites := make(map[uint64]*teststruct.TestSuite, 0)

		for _, v := range testsIDs {

			testsID, _ := strToTestsIDs(v)

			kv, err := im.ConsulClient.GetKey(v)
			if err != nil {
				return err
			}
			// tstepAddr is "TSid/TCid/TStepid"
			tstepAddr := kv.Value

		}
	*/
	return nil
}

type testSchedulerJSON struct {
	value map[string][]string
}

func (tsch *testScheduler) UnmarshalJSON(data []byte) error {

	var jsonTScheduler testSchedulerJSON
	if errjs := json.Unmarshal(data, &jsonTScheduler); errjs != nil {
		return errjs
	}

	var err error
	*tsch, err = jsonTScheduler.toTSched()
	if err != nil {
		return err
	}

	return nil
}

func (jtsh testSchedulerJSON) toTSched() (testScheduler, error) {

	var tsch testScheduler

	tsch.value = make(map[string][]testID, len(jtsh.value))

	for k, v := range jtsh.value {

		tsch.value[k] = nil
		var temp []testID
		for _, w := range v {

			tid, err := strToTestsIDs(w)

			if err != nil {
				return testScheduler{}, err
			}
			temp = append(temp, tid)

		}

	}

	return tsch, nil

}

func strToTestsIDs(s string) (testID, error) {

	raw := strings.Split(s, "/")

	tsid, err := strconv.ParseUint(raw[0], 10, 64)
	if err != nil {
		return testID{}, err
	}
	tcid, err := strconv.ParseUint(raw[1], 10, 64)
	if err != nil {
		return testID{}, err
	}
	tstepid, err := strconv.ParseUint(raw[2], 10, 64)
	if err != nil {
		return testID{}, err
	}

	tid := testID{
		TestSuite: tsid,
		TestCase:  tcid,
		TestStep:  tstepid,
	}
	return tid, nil
}

type testScheduler struct {
	value map[string][]testID
}

type testID struct {
	TestSuite uint64
	TestCase  uint64
	TestStep  uint64
}
