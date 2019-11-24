package hash

import (
	"fmt"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/vigie"
	"reflect"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
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
func TestExecutor_Validate(t *testing.T) {

	// TS01 Step all Algo
	TS01_empty := vigie.TestStep{"Name": "e1", "Algo": "piedpiper", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_crazyalgo := vigie.TestStep{"Name": "e1", "Algo": "", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_md5 := vigie.TestStep{"Name": "e1", "Algo": "md5", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_sha1 := vigie.TestStep{"Name": "e1", "Algo": "sha1", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_sha2 := vigie.TestStep{"Name": "e1", "Algo": "sha2", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_sha256 := vigie.TestStep{"Name": "e1", "Algo": "sha256", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_sha512 := vigie.TestStep{"Name": "e1", "Algo": "sha512", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_blake256 := vigie.TestStep{"Name": "e1", "Algo": "blake2b-256", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_blake384 := vigie.TestStep{"Name": "e1", "Algo": "blake2b-384", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS01_blake512 := vigie.TestStep{"Name": "e1", "Algo": "blake2b-512", "URL": "http://localhost/foo.txt", "Interval": 1000}
	// TS02 Step Algo Case
	TS02_md5 := vigie.TestStep{"Algo": "MD5", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS02_sha1 := vigie.TestStep{"Algo": "SHA1", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS02_sha2 := vigie.TestStep{"Algo": "Sha2", "URL": "http://localhost/foo.txt", "Interval": 1000}
	TS02_blake256 := vigie.TestStep{"Name": "e1", "Algo": "Blake2B-256", "URL": "http://localhost/foo.txt", "Interval": 1000}
	// TS03 URL
	TS03_1 := vigie.TestStep{"Algo": "md5", "URL": "1.1.1.1", "Interval": 1000}
	TS03_2 := vigie.TestStep{"Algo": "md5", "URL": "vincoll.io", "Interval": 1000}
	TS03_3 := vigie.TestStep{"Algo": "md5", "URL": "https://foo.bar", "Interval": 1000}
	TS03_4 := vigie.TestStep{"Name": "e1", "Algo": "md5", "URL": "https://www.url.x.lan:12345/url/path/foo?bar=zzz#user", "Interval": 1000}
	// TS03 URL Err
	TS03_5 := vigie.TestStep{"Algo": "md5", "URL": "", "Interval": 1000}
	TS03_6 := vigie.TestStep{"Algo": "md5", "URL": "/vincoll.io", "Interval": 1000}
	TS03_7 := vigie.TestStep{"Algo": "md5", "URL": "//foo.bar", "Interval": 1000}
	TS03_8 := vigie.TestStep{"Name": "e1", "Algo": "md5", "URL": "localhost", "Interval": 1000}
	type args struct {
		step *teststruct.Step
	}

	tests := []struct {
		name    string
		e       Executor
		args    args
		wantErr bool
	}{
		// Step Algos
		{name: "TS01_empty", e: Executor{}, args: args{step: &TS01_empty}, wantErr: true},
		{name: "TS01_crazyalgo", e: Executor{}, args: args{step: &TS01_crazyalgo}, wantErr: true},
		{name: "TS01_md5", e: Executor{}, args: args{step: &TS01_md5}, wantErr: false},
		{name: "TS01_sha1", e: Executor{}, args: args{step: &TS01_sha1}, wantErr: false},
		{name: "TS01_sha2", e: Executor{}, args: args{step: &TS01_sha2}, wantErr: false},
		{name: "TS01_sha256", e: Executor{}, args: args{step: &TS01_sha256}, wantErr: false},
		{name: "TS01_sha512", e: Executor{}, args: args{step: &TS01_sha512}, wantErr: false},
		{name: "TS01_blake256", e: Executor{}, args: args{step: &TS01_blake256}, wantErr: true},
		{name: "TS01_blake384", e: Executor{}, args: args{step: &TS01_blake384}, wantErr: true},
		{name: "TS01_blake512", e: Executor{}, args: args{step: &TS01_blake512}, wantErr: true},
		// Step Algo String Case
		{name: "TS02_md5", e: Executor{}, args: args{step: &TS02_md5}, wantErr: false},
		{name: "TS02_sha1", e: Executor{}, args: args{step: &TS02_sha1}, wantErr: false},
		{name: "TS02_sha2", e: Executor{}, args: args{step: &TS02_sha2}, wantErr: false},
		{name: "TS02_blake256", e: Executor{}, args: args{step: &TS02_blake256}, wantErr: true},
		// Step URL
		{name: "TS03_1", e: Executor{}, args: args{step: &TS03_1}, wantErr: false},
		{name: "TS03_2", e: Executor{}, args: args{step: &TS03_2}, wantErr: false},
		{name: "TS03_3", e: Executor{}, args: args{step: &TS03_3}, wantErr: false},
		{name: "TS03_4", e: Executor{}, args: args{step: &TS03_4}, wantErr: false},
		{name: "TS03_5", e: Executor{}, args: args{step: &TS03_5}, wantErr: true},
		{name: "TS03_6", e: Executor{}, args: args{step: &TS03_6}, wantErr: true},
		{name: "TS03_7", e: Executor{}, args: args{step: &TS03_7}, wantErr: true},
		{name: "TS03_8", e: Executor{}, args: args{step: &TS03_8}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Validate(tt.args.step); (err != nil) != tt.wantErr {
				t.Errorf("Probe.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutor_GetDefaultAssertions(t *testing.T) {
	tests := []struct {
		name string
		e    Executor
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

func TestExecutor_Run(t *testing.T) {
	// TS01
	TS01_md5 := vigie.TestStep{"Algo": "md5", "URL": "http://localhost:6080/robots.txt", "Interval": 1000}
	TS01_md5Wait1sec := vigie.TestStep{"Algo": "md5", "URL": "http://localhost:6080/delay/1", "Interval": 1000}

	//	GOT := teststruct.ExecutorResult{"result.executor.__len__":4, "result.executor":hash.Probe{Name:"", Algo:"md5", URL:"http://localhost:6080/robots.txt", Interval:1000}, "result.executor.interval":1000, "result.err":"", "result.hash":"46fd03688e49c6b6b0a2b7d3553c1e42", "result.executor.url":"http://localhost:6080/robots.txt", "result.success":"true", "__type__":"ResultStatus", "result.__len__":4, "result":hash.ResultStatus{Probe:hash.Probe{Name:"", Algo:"md5", URL:"http://localhost:6080/robots.txt", Interval:1000}, ResultStatus:"true", Err:"", Hash:"46fd03688e49c6b6b0a2b7d3553c1e42"}, "result.executor.__type__":"Probe", "result.executor.name":"", "result.executor.algo":"md5"}
	logctx := log.WithField("", "")

	type args struct {
		l    vigie.Logger
		step teststruct.Step
	}
	tests := []struct {
		name    string
		e       Executor
		args    args
		want    teststruct.ExecutorResult
		wantErr bool
	}{
		{name: "TS01_md5", e: Executor{}, args: args{l: logctx, step: TS01_md5}, want: nil, wantErr: false},
		{name: "TS01_md5Wait1sec", e: Executor{}, args: args{l: logctx, step: TS01_md5Wait1sec}, want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Run(tt.args.l, tt.args.step)
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe.Start() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("_____")
			fmt.Printf("%#v", got)
			fmt.Println("_____")

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Probe.Start() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_Hash(t *testing.T) {
	tests := []struct {
		name    string
		e       *Executor
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Hash()
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe.Hash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Probe.Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloadFromUrl(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "TS01_robot", args: args{url: "http://localhost:6080/robot.txt"}, want: "httplocalhost6080robot.txt", wantErr: false},
		{name: "TS01_Delay1sec", args: args{url: "http://localhost:6080/delay/1"}, want: "httplocalhost6080delay1", wantErr: false},
		{name: "TS01_NoDirectFile", args: args{url: "http://localhost:6080"}, want: "httplocalhost6080", wantErr: false},

		// Err
		{name: "TS02_InvalidPort", args: args{url: "http://localhost:999999"}, wantErr: true},
		{name: "TS02_none", args: args{url: "http://localhost:49152"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := downloadFromUrl(tt.args.url)

			if (err != nil) != tt.wantErr {
				t.Errorf("downloadFromUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {

				// sourceFile Name after DL:
				tokens := strings.Split(tt.args.url, "/")
				fileName := tokens[len(tokens)-1]
				url2file := strings.NewReplacer(":", "", "/", "", "\\\\", "")
				fileName = url2file.Replace(fileName)

				if !strings.Contains(got, fileName) {
					t.Errorf("Not OK")
					return
				}
			}
		})
	}
}
