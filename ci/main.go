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
	Sec.GH_TOKEN = os.Getenv("GH_TOKEN")
	if Sec.GH_TOKEN == "" {
		log.Fatal("GH_TOKEN is not set")
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
	////////////////////////////////////////////////////////////

	// get host directory
	project := client.Host().Directory(".")

	fmt.Println("Docker multistage build...")
	// build app
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

	// publish binary on alpine base
	finalStage := client.Container().
		From("alpine").
		WithLabel("org.opencontainers.image.title", "vigie").
		WithLabel("org.opencontainers.image.version", Help.ShaShort).
		WithLabel("org.opencontainers.image.created", Help.DateRFC3339).
		WithFile("/vigie", builderStage.File("/app/vigie")).
		WithExec([]string{"mkdir", "--parents", "/app/config"}).
		WithEntrypoint([]string{"/vigie"}).
		WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{Args: []string{"version"}})

	// Publish to Registry

	finalStage = finalStage.WithRegistryAuth("ghcr.io", "vincoll", client.SetSecret("gh_token", Sec.GH_TOKEN))

	tags := []string{Help.ShaShort, "latest"}
	for _, tag := range tags {

		addr, err := finalStage.Publish(ctx, fmt.Sprintf("ghcr.io/%s/vigie:%s", "vincoll", tag))
		if err != nil {
			panic(fmt.Errorf("failed to publish image: %w", err))
		}
		fmt.Printf("Published image to :%s\n", addr)
	}

	fmt.Println(finalStage)

}

func buildImage(ctx context.Context, client *dagger.Client) (error, []*dagger.Container) {
	fmt.Println("Building with Dagger")

	// Dagger https://github.com/dagger/dagger/issues/4567
	return nil, nil
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
	GH_TOKEN string
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

//
/*

https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126


*/
