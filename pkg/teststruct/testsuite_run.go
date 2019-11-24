package teststruct

type TSDescribe struct {
	Errors         int          `json:"errors"`
	Status         bool         `json:"result"`
	Failures       int          `json:"failures"`
	Name           string       `json:"name"`
	ID             int64        `json:"id"`
	TestCaseCount  int          `json:"testcasescount"`
	TestStepsCount int          `json:"teststepscount"`
	TestCases      []TCDescribe `json:"testcases"`
}

func (ts *TestSuite) ToJSON() *TSDescribe {

	ts.Mutex.RLock()

	var TSDesc TSDescribe

	TSDesc.TestCases = make([]TCDescribe, 0, len(ts.TestCases))
	for _, tc := range ts.TestCases {
		TSDesc.TestStepsCount += tc.countTestStep()
		tsD := tc.ToJSON()
		TSDesc.TestCases = append(TSDesc.TestCases, *tsD)
	}

	TSDesc.Name = ts.Name
	TSDesc.ID = ts.ID
	TSDesc.TestCaseCount = len(ts.TestCases)
	TSDesc.Failures = ts.Failures()
	TSDesc.Status = ts.isSuccess()

	ts.Mutex.RUnlock()

	return &TSDesc
}

func (ts *TestSuite) UpdateStatus() {
	ts.Mutex.Lock()
	defer ts.Mutex.Unlock()

	for _, tc := range ts.TestCases {
		tc.Mutex.RLock()

		if tc.Status == Failure {
			ts.Status = Failure
			tc.Mutex.RUnlock()
			return
		}
		tc.Mutex.RUnlock()
	}
	ts.Status = Success
	return
}

func (ts *TestSuite) SetStatus(status StepStatus) {
	ts.Mutex.Lock()
	ts.Status = status
	ts.Mutex.Unlock()
}

func (ts *TestSuite) isSuccess() bool {
	ts.Mutex.RLock()
	defer ts.Mutex.RUnlock()
	for _, tc := range ts.TestCases {
		tc.Mutex.RLock()
		//if tc.Details() > 0 || tc.Details() > 0 {
		if tc.Status != Success {
			tc.Mutex.RUnlock()
			return false
		}
		tc.Mutex.RUnlock()
	}
	return true
}

func (ts *TestSuite) Failures() (count int) {
	ts.Mutex.RLock()
	for _, tc := range ts.TestCases {
		tc.Mutex.RLock()
		count += tc.FailureCount()
		tc.Mutex.RUnlock()
	}
	ts.Mutex.RUnlock()
	return count
}
