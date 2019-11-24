package teststruct



type TSHeader struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (ts *TestSuite) ToHeader() TSHeader {

	ts.Mutex.RLock()

	hts := TSHeader{
		Name:   ts.Name,
		ID:     ts.ID,
		Status: ts.Status.String(),
	}

	ts.Mutex.RUnlock()

	return hts

}

type TSAlertShort struct {
	TSHeader
	TestCases map[int64]TCAlertShort `json:"testcases"`
}

func (ts *TestSuite) ToAlertShortTSRec() TSAlertShort {

	htc := ts.ToHeader()

	AsTC := TSAlertShort{
		TSHeader:  htc,
		TestCases: nil,
	}

	return AsTC
}

func (ts *TestSuite) ToAlertShortTS() TSAlertShort {

	htc := ts.ToHeader()

	TCAlerts := make(map[int64]TCAlertShort, 0)

	AsTC := TSAlertShort{
		TSHeader:  htc,
		TestCases: TCAlerts,
	}

	/*
		ts.Mutex.RLock()
		for _, tc := range ts.TestCases {
			TCAlerts = append(TCAlerts, tc.ToAlertShortTC())
		}
		ts.Mutex.RUnlock()
	*/
	return AsTC
}
