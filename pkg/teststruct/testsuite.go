package teststruct

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/utils"
)

type JSONTestSuite struct {
	Config        configTestStructJson `json:"config"`
	Name          string               `json:"name"`
	JsonTestCases []JSONTestCase       `json:"testcases"`
	Vars          map[string][]string  `json:"vars"`
}

type TestSuite struct {
	Mutex       sync.RWMutex
	Name        string
	ID          int64
	Status      StepStatus
	TestCases   map[int64]*TestCase
	LastChange  time.Time
	CountChange uint
	SourceFile  string // Public
}

// UnmarshalJSON Interface for TestSuite_OLD struct
// This Go interface allow to **
// importing deeply TestSuites + TC + Tsteps
// This function is call by unmarshalTestSuiteFile in
// TC will also use a UnmarshalJSON Interface
// TStep

func (ts *TestSuite) UnmarshalJSON(data []byte) error {
	var jsonTS JSONTestSuite
	if errjs := json.Unmarshal(data, &jsonTS); errjs != nil {
		return errjs
	}

	var err error
	*ts, err = jsonTS.toTestSuite()
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"testsuite": jsonTS.Name,
		}).Fatalf("A TestSuite is invalid: %s", err)

		return err
	}
	return nil
}

// toTestSuite convert JsonTestSuite into TestSuite after Validations
func (jts JSONTestSuite) toTestSuite() (TestSuite, error) {

	// TestSuite init
	var ts TestSuite

	if jts.Name == "" {
		return ts, fmt.Errorf("name is missing or empty")
	}
	ts.Name = jts.Name

	if jts.JsonTestCases == nil {
		return TestSuite{}, fmt.Errorf("no testcase detected in %q testsuite", jts.Name)
	}

	// UnMarshall spécifique de la config avec conversion de strings type (1d,7m) en time.duration
	ctsTS, err := unmarshallConfigTestStruct(jts.Config)
	if err != nil {

		utils.Log.WithFields(logrus.Fields{
			"testsuite": ts.Name,
		}).Fatalf("config declaration : %s", err)

		return TestSuite{}, fmt.Errorf("config declaration: %s", err)
	}

	ts.TestCases = make(map[int64]*TestCase, len(jts.JsonTestCases))

	mergeMapsTS := utils.MergeMaps(utils.ALLVARS, jts.Vars)

	// Apply config inheritance TestSuite => TC
	for i, jtc := range jts.JsonTestCases {

		testcase, jtcErr := jtc.toTestCase(&ctsTS, mergeMapsTS)
		if jtcErr != nil {

			utils.Log.WithFields(logrus.Fields{
				"testsuite": ts.Name,
			}).Fatalf("Cannot import : %s", jtcErr)

		}

		i64 := int64(i)
		testcase.ID = i64
		ts.TestCases[i64] = &testcase

	}

	return ts, nil
}

/*
// GetTestcaseByID get a TestCase by is ID in the current TestSuite.
func (ts *TestSuite) GetTestcaseByID(TCid int64) (tc *TestCase) {

	ts.Mutex.RLock()
	tc = ts.TestCases[TCid]
	ts.Mutex.RUnlock()
	return tc

}
*/

func (ts *TestSuite) WithoutTC() TestSuite {

	tsBis := *ts
	// Reset TestCase
	tsBis.TestCases = make(map[int64]*TestCase, 1)
	tsBis.Mutex = sync.RWMutex{}
	return tsBis

}
