package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/sethvargo/go-envconfig"
)

func init() {

	ctx := context.Background()

	// Github Envs

	var c githubContext
	if err := envconfig.Process(ctx, &c); err != nil {
		log.Fatal(err)
	}

}

///
/// Build Context
///

type BuildContext struct {
	DateRFC3339  string
	Sha          string
	ShaShort     string
	Version      string
	Architecture string
}

func NewBuildContext() BuildContext {

	// Build Context
	bc := BuildContext{
		Sha:          getSHA(),
		ShaShort:     getSHA()[0:7],
		DateRFC3339:  time.Now().Format(time.RFC3339),
		Architecture: fmt.Sprintf("linux/%s", runtime.GOARCH),
	}

	return bc
}


///
/// GitHub Context
///

var GithubContext githubContext

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

///
/// Tools Functions
///

// getSHA returns the short and long sha of the current git commit
func getSHA() (longSha string) {

	cmd, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Println("Error getting SHA:", err)
		os.Exit(1)
	}
	if len(cmd) == 0 {
		fmt.Println("Error getting SHA: no output")
		os.Exit(1)
	}

	return string(cmd)
}
