package vigiescheduler

import (
	"reflect"
	"testing"

	"github.com/vincoll/vigie/internal/api/conf"
)

func Test_loadVigieConfigFile(t *testing.T) {
	type args struct {
		confpath string
	}
	tests := []struct {
		name   string
		args   args
		wantVc conf.VigieAPIConf
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotVc := loadVigieConfigFile(tt.args.confpath); !reflect.DeepEqual(gotVc, tt.wantVc) {
				t.Errorf("loadVigieConfigFile() = %v, want %v", gotVc, tt.wantVc)
			}
		})
	}
}

func Test_applyEnvironment(t *testing.T) {
	type args struct {
		vc *conf.VigieAPIConf
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{&conf.VigieAPIConf{Environment: ""}}, want: "production"},
		{name: "dev", args: args{&conf.VigieAPIConf{Environment: "dev"}}, want: "development"},
		{name: "develop", args: args{&conf.VigieAPIConf{Environment: "develop"}}, want: "development"},
		{name: "development", args: args{&conf.VigieAPIConf{Environment: "development"}}, want: "development"},
		{name: "dev space", args: args{&conf.VigieAPIConf{Environment: "dev "}}, want: "production"},
		{name: "wrong", args: args{&conf.VigieAPIConf{Environment: "wrong"}}, want: "production"},
		{name: "production", args: args{&conf.VigieAPIConf{Environment: "production"}}, want: "production"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyEnvironment(tt.args.vc); got != tt.want {
				t.Errorf("applyEnvironment() = %v, want %v", got, tt.want)
			}
		})
	}
}
