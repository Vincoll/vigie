package testgrp

import (
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"google.golang.org/protobuf/types/known/anypb"
)

// VigieTestREST is the expected struct received by the REST API
// It's a less rigid struct that the full ProbeComplete Protobuf
type VigieTestREST struct {
	Metadata   *probe.Metadata        `json:"metadata"`
	Spec       *anypb.Any             `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

func (vtr *VigieTestREST) toVigieTest() *probemgmt.VigieTest {

	vt := probemgmt.VigieTest{
		Metadata:   vtr.Metadata,
		Spec:       vtr.Spec,
		Assertions: vtr.Assertions,
	}

	return &vt
}

/*
// VigieTestJSON is a transition struct for UnMarshalJSON
// to Init and Validate data comming from the REST API
type VigieTestJSON struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       interface{}            `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

// UnmarshalJSON vigieTest in a "Temp" Struct much closer than the REST Payload will be
// Conversion will be made to generate a clean VigieTestREST
func (vtr *VigieTestREST) UnmarshalJSON(data []byte) error {

	var jsonTS VigieTestJSON
	if errjs := json.Unmarshal(data, &jsonTS); errjs != nil {
		return errjs
	}
	/*
		var pc probe.ProbeComplete
		x := protojson.Unmarshal(data, &pc)
		print(x)

	var err error

	if err != nil {
		return fmt.Errorf("VigieTestREST is invalid: %s", err)
	}
	return nil
}

*/
