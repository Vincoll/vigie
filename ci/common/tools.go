package common

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"dagger.io/dagger"
)

var BuildContext buildContext
var GithubContext githubContext

type buildContext struct {
	DateRFC3339  string
	Sha          string
	ShaShort     string
	Version      string
	Architecture string
}

type githubContext struct {
	Workflow   string `env:"GITHUB_WORKFLOW"`
	RunID      string `env:"GITHUB_RUN_ID"`
	RunNumber  string `env:"GITHUB_RUN_NUMBER"`
	Action     string `env:"GITHUB_ACTION"`
	Actions    string `env:"GITHUB_ACTIONS"`
	Actor      string `env:"GITHUB_ACTOR"`
	Repository string `env:"GITHUB_REPOSITORY"`
	EventName  string `env:"GITHUB_EVENT_NAME"`
	EventPath  string `env:"GITHUB_EVENT_PATH"`
	Workspace  string `env:"GITHUB_WORKSPACE"`
	SHA        string `env:"GITHUB_SHA"`
	Ref        string `env:"GITHUB_REF"`
	HeadRef    string `env:"GITHUB_HEAD_REF"`
	BaseRef    string `env:"GITHUB_BASE_REF"`
	ServerURL  string `env:"GITHUB_SERVER_URL"`
	APIURL     string `env:"GITHUB_API_URL"`
	GraphQLURL string `env:"GITHUB_GRAPHQL_URL"`
}

func init() {

	var bc buildContext

	// Vars
	bc.ShaShort, bc.Sha = getSHA()
	bc.DateRFC3339 = time.Now().Format(time.RFC3339)
	bc.Version = "0.0.1"
	bc.Architecture = "" //dagger.Platform(fmt.Sprintf("linux/%s", runtime.GOARCH))

	BuildContext = bc

}

// getSHA returns the short and long sha of the current git commit
func getSHA() (shortSha string, longSha string) {

	cmd, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Println("Error getting SHA:", err)
		os.Exit(1)
	}
	if len(cmd) == 0 {
		fmt.Println("Error getting SHA: no output")
		os.Exit(1)
	}

	sha := string(cmd)
	return sha[0:7], sha
}

