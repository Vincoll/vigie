package testgrp

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	*/
	var err error
	*vtr, err = jsonTS.toVigieTest()
	if err != nil {
		return fmt.Errorf("VigieTestREST is invalid: %s", err)
	}
	return nil
}

func (jvt *VigieTestJSON) toVigieTest() (VigieTestREST, error) {

	// VigieTestREST init
	var vt VigieTestREST

	//
	// Metadata
	//
	if jvt.Metadata.Name == "" {
		return VigieTestREST{}, fmt.Errorf("name is missing or empty")
	}
	vt.Metadata = &jvt.Metadata

	//
	// Spec - Validation and Probe Init with default values
	//

	var message proto.Message
	// "This list will be registered elsewhere":
	// This must be re-factored
	switch jvt.Metadata.Type {
	case "icmp":
		m := icmp.New()

		x, err := json.Marshal(jvt.Spec)
		if err != nil {
			return VigieTestREST{}, err
		}
		// Populate the ICMP Spec with the data from the REST Payload
		err = protojson.Unmarshal(x, m)
		if err != nil {
			return VigieTestREST{}, err
		}
		// Validate and Init the Probe
		err = m.ValidateAndInit()
		if err != nil {
			return VigieTestREST{}, err
		}
		message = m
	case "tcp":
		message = &icmp.Probe{}
	default:
		return VigieTestREST{}, fmt.Errorf("type %q is invalid", jvt.Metadata.Type)
	}

	// convert vt.Spec to anypb.any
	spec, err := anypb.New(message)
	if err != nil {
		return VigieTestREST{}, err
	}
	spec.TypeUrl = jvt.Metadata.Type

	vt.Spec = spec

	//
	// Assertions
	//

	vt.Assertions = jvt.Assertions

	return vt, nil
}
