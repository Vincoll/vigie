package importing

import (
	"reflect"
	"testing"

	"github.com/vincoll/vigie/pkg/assertion"
)

func Test_initAssert(t *testing.T) {
	type args struct {
		rawAssert string
	}

	equalFoo := assertion.Assert{Key: "asw", Method: assertion.Equal, Value: "foo"}

	tests := []struct {
		name    string
		args    args
		want    []assertion.Assert
		wantErr bool
	}{
		{name: "string foo", args: args{rawAssert: "asw == foo"}, want: []assertion.Assert{equalFoo}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initAssert(tt.args.rawAssert)
			if (err != nil) != tt.wantErr {
				t.Errorf("initAssert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initAssert() = %v, want %v", got, tt.want)
			}
		})
	}
}
