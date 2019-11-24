package teststruct

import (
	"reflect"
	"testing"
)

func Test_unmarshallconfig(t *testing.T) {
	type args struct {
		ctjson configTestStructJson
	}
	tests := []struct {
		name    string
		args    args
		want    configTestStruct
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshallConfigTestStruct(tt.args.ctjson)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshallConfigTestStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshallConfigTestStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
