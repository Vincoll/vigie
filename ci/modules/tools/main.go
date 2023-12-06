package main

import (
	"context"
	"log"

	"github.com/sethvargo/go-envconfig"
)

// https://docs.github.com/en/actions/learn-github-actions/variables
type GithubContext struct {
	Workflow        string `env:"GITHUB_WORKFLOW"`
	RunID           string `env:"GITHUB_RUN_ID"`
	RunNumber       string `env:"GITHUB_RUN_NUMBER"`
	Action          string `env:"GITHUB_ACTION"`
	ActionPath      string `env:"GITHUB_ACTION_PATH"`
	Actor           string `env:"GITHUB_ACTOR"`
	ActorId         string `env:"GITHUB_ACTOR_ID"`
	Repository      string `env:"GITHUB_REPOSITORY"`
	RepositoryOwner string `env:"GITHUB_REPOSITORY_OWNER"`
	RepositoryId    string `env:"GITHUB_REPOSITORY_ID"`
	EventName       string `env:"GITHUB_EVENT_NAME"`
	EventPath       string `env:"GITHUB_EVENT_PATH"`
	Workspace       string `env:"GITHUB_WORKSPACE"`
	SHA             string `env:"GITHUB_SHA"`
	Ref             string `env:"GITHUB_REF"`
	HeadRef         string `env:"GITHUB_HEAD_REF"`
	BaseRef         string `env:"GITHUB_BASE_REF"`
	ServerURL       string `env:"GITHUB_SERVER_URL"`
	APIURL          string `env:"GITHUB_API_URL"`
	GraphQLURL      string `env:"GITHUB_GRAPHQL_URL"`
	RunAttempt      string `env:"GITHUB_RUN_ATTEMPT"`
	RunnerOS        string `env:"RUNNER_OS"`
	RunnerArch      string `env:"RUNNER_ARCH"`
	RunnerTemp      string `env:"RUNNER_TEMP"`
}

var GH GithubContext

func init() {

	// Env
	if err := envconfig.Process(context.Background(), &GH); err != nil {
		log.Fatal(err)
	}

}
