package assertion

import (
	"testing"

	"github.com/smartystreets/assertions"

	"github.com/vincoll/vigie/pkg/teststruct"
)

func Test_assert(t *testing.T) {

	type args struct {
		probeValue string
		tAssert    *assertion.Assert
	}
	tests := []struct {
		name     string
		args     args
		wantRes  string
		wantFail string
	}{
		{
			name:     "a == a",
			args:     args{probeValue: "a", tAssert: &assertion.Assert{RawVerb: "==", Verb: assertions.ShouldEqual, Value: "a"}},
			wantFail: "",
			wantRes:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, gotFail := assert(tt.args.probeValue, tt.args.tAssert)
			if gotRes != tt.wantRes {
				t.Errorf("assert() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
			if gotFail != tt.wantFail {
				t.Errorf("assert() gotFail = %v, want %v", gotFail, tt.wantFail)
			}
		})
	}
}
