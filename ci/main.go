package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"dagger.io/dagger"
)

func init() {

	// Helpers
	Help.ShaShort, Help.Sha = getSHA()
	Help.DateRFC3339 = time.Now().Format(time.RFC3339)
	Help.Version = "0.0.1"

	// Secrets
	Sec.GITHUB_TOKEN = os.Getenv("GITHUB_TOKEN")
	if Sec.GITHUB_TOKEN == "" {
		log.Fatal("Env Var GITHUB_TOKEN is not set")
	}

}

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	// Build Docker Multistage ---
	ctnrs, err := buildImage(ctx, client)
	if err != nil {
		panic(fmt.Errorf("failed to build docker image: %w", err))
	}

	for _, ctnr := range ctnrs {

		tags := []string{Help.ShaShort, "latest"}

		err := publishImage(ctx, client, ctnr, tags)
		if err != nil {
			panic(fmt.Errorf("failed to publish image: %w", err))
		}

	}
	fmt.Println("Done")
}

// buildImage builds a docker image with a multistage build
func buildImage(ctx context.Context, client *dagger.Client) ([]*dagger.Container, error) {

	fmt.Println("Docker multistage build...")
	ctnrs := []*dagger.Container{}

	project := client.Host().Directory(".")

	builderStage := client.Container().
		From(fmt.Sprintf("golang:%s-alpine", "1.21")).
		WithEnvVariable("CGO_ENABLED", "0").
		WithWorkdir("/app").
		// run `go mod download` with only go.mod files (re-run only if mod files have changed)
		WithDirectory(".", project, dagger.ContainerWithDirectoryOpts{
			Include: []string{"**/go.mod", "**/go.sum"},
		}).
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
		WithExec([]string{"go", "mod", "download"}).
		// run `go build` with all source
		WithMountedDirectory(".", project).
		WithExec([]string{"go", "build",
			"-ldflags",
			"-X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=" + Help.ShaShort + " " +
				"-X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=" + Help.DateRFC3339 + " " +
				"-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=" + Help.Version + " ",
			"-o", "vigie"}).
		// include a cache for go build
		WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build"))

	finalStage := client.Container().
		From("alpine:latest").
		WithLabel("org.opencontainers.image.title", "vigie").
		WithLabel("org.opencontainers.image.version", Help.ShaShort).
		WithLabel("org.opencontainers.image.created", Help.DateRFC3339).
		WithFile("/vigie", builderStage.File("/app/vigie")).
		WithExec([]string{"mkdir", "--parents", "/app/config"}).
		WithEntrypoint([]string{"/vigie"}).
		WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{Args: []string{"version"}})

	ctnrs = append(ctnrs, finalStage)

	return ctnrs, nil
}

func publishImage(ctx context.Context, client *dagger.Client, ctnr *dagger.Container, tags []string) error {

	// Publish to Registry ---
	ctnr = ctnr.WithRegistryAuth("ghcr.io", "vincoll", client.SetSecret("gh_token", Sec.GITHUB_TOKEN))
	for _, tag := range tags {
		addr, err := ctnr.Publish(ctx, fmt.Sprintf("ghcr.io/%s/vigie:%s", "vincoll", tag))
		if err != nil {
			return err
		}
		fmt.Printf("published image to :%s\n", addr)
	}
	return nil
}

var Help Helpers

type Helpers struct {
	DateRFC3339 string
	Sha         string
	ShaShort    string
	Version     string
}

var Sec Secrets

type Secrets struct {
	GITHUB_TOKEN string
}

func getSHA() (shortSha string, longSha string) {

	cmd, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Println("Error getting SHA:", err)
		os.Exit(1)
	}
	sha := string(cmd)
	return sha[0:7], sha
}

/*

https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126
	// Dagger https://github.com/dagger/dagger/issues/4567


*/
