package teststruct

import (
	"time"

	"github.com/vincoll/vigie/pkg/utils"
)

// unmarshallConfigTestStruct convert raw json data into more easily manipulated types
func unmarshallConfigTestStruct(ctjson configTestStructJson) (configTestStruct, error) {

	// Convert string duration format (1d, 127ms...) to time.duration
	cts := configTestStruct{
		Concurrency: ctjson.Concurrency,
		Retry:       ctjson.Retry,
		Frequency:   map[string]time.Duration{},
		Retrydelay:  map[string]time.Duration{},
		Timeout:     map[string]time.Duration{},
	}

	for k, v := range ctjson.Frequency {
		dur, err := utils.ParseDuration(v)
		if err != nil {
			return cts, err
		}
		cts.Frequency[k] = dur
	}

	for k, v := range ctjson.Timeout {
		dur, err := utils.ParseDuration(v)
		if err != nil {
			return cts, err
		}
		cts.Timeout[k] = dur
	}

	for k, v := range ctjson.Retrydelay {
		dur, err := utils.ParseDuration(v)
		if err != nil {
			return cts, err
		}
		cts.Retrydelay[k] = dur
	}

	return cts, nil
}
