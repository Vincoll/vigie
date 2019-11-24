package icmp

import (
	"github.com/vincoll/vigie/pkg/vigie"
	"reflect"
	"testing"

	ping "github.com/sparrc/go-ping"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want teststruct.Executor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVigie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_GetName(t *testing.T) {
	tests := []struct {
		name string
		e    Probe
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetName(); got != tt.want {
				t.Errorf("Probe.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_GetDefaultAssertions(t *testing.T) {
	tests := []struct {
		name string
		e    Probe
		want teststruct.StepAssertions
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetDefaultAssertions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Probe.GetDefaultAssertions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_Validate(t *testing.T) {
	type args struct {
		step *teststruct.Step
	}
	tests := []struct {
		name    string
		e       Probe
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Validate(tt.args.step); (err != nil) != tt.wantErr {
				t.Errorf("Probe.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutor_Run(t *testing.T) {
	type args struct {
		l    vigie.Logger
		step teststruct.Step
	}
	tests := []struct {
		name    string
		e       Probe
		args    args
		want    teststruct.ExecutorResult
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Run(tt.args.l, tt.args.step)
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Probe.Start() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_pingToHost(t *testing.T) {
	tests := []struct {
		name    string
		e       *Probe
		want    *ping.Statistics
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.process()
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe.process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Probe.process() = %v, want %v", got, tt.want)
			}
		})
	}
}
