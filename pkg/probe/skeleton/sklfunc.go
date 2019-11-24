package skeleton

import (
	"strconv"
	"time"

	"github.com/vincoll/vigie/pkg/probe"
)

func func1(v string) ProbeAnswer {

	start := time.Now()
	returnedAnswer, err := strconv.Atoi(v)
	elapsed := time.Since(start)

	// Error
	if err != nil {
		pi := probe.ProbeInfo{
			Status: -3,
			Error:  err.Error(),
		}
		// Define Vigie ProbeCode Error
		switch err.Error() {
		case "Classic Error":
			pi.ProbeCode = 666
		}

		pa := ProbeAnswer{
			ResponseTime: elapsed.Seconds(),
			ProbeInfo:    pi,
		}
		return pa
	}

	// Success
	pi := probe.ProbeInfo{
		Status: 1,
	}

	pa := ProbeAnswer{
		Answer:       returnedAnswer,
		ResponseTime: elapsed.Seconds(),
		ProbeInfo:    pi,
	}

	return pa
}
