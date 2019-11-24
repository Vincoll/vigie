package debug

import "github.com/vincoll/vigie/pkg/probe"

func (p *Probe) genSuccess() ProbeAnswer {

	pi := probe.ProbeInfo{
		Error:        "",
		Status:       probe.Success,
		ProbeCode:    0,
		ResponseTime: 66,
	}

	pa := ProbeAnswer{
		Answer:    p.Answer,
		ProbeInfo: pi,
	}

	return pa
}

func (p *Probe) genSuccessMsg(msg string) ProbeAnswer {

	pi := probe.ProbeInfo{
		Error:        "",
		Status:       probe.Success,
		ProbeCode:    0,
		ResponseTime: 66,
	}

	pa := ProbeAnswer{
		Answer:    msg,
		ProbeInfo: pi,
	}

	return pa
}

func (p *Probe) genTimeout() ProbeAnswer {

	pi := probe.ProbeInfo{
		Error:        "Probe exec timeout",
		Status:       probe.Timeout,
		ProbeCode:    0,
		ResponseTime: 666,
	}

	pa := ProbeAnswer{
		Answer:    "",
		ProbeInfo: pi,
	}

	return pa
}

func (p *Probe) genError() ProbeAnswer {

	pi := probe.ProbeInfo{
		Error:     "Probe exec error",
		Status:    probe.Error,
		ProbeCode: p.ErrorCode,
	}

	pa := ProbeAnswer{
		Answer:    "",
		ProbeInfo: pi,
	}

	return pa
}
