package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"dagger.io/dagger"
	platformFormat "github.com/containerd/containerd/platforms"
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

	// ---

	var archs = []dagger.Platform{"linux/amd64", "linux/arm64"}
	platforms, err := buildImage(ctx, client, archs)
	if err != nil {
		panic(fmt.Errorf("failed to build docker image: %w", err))
	}

	tags := []string{Help.ShaShort, "latest"}
	err = publishImage(ctx, client, platforms, tags)
	if err != nil {
		panic(fmt.Errorf("failed to publish image: %w", err))
	}

	fmt.Println("Done")
}

// buildImage builds a docker image with a multistage build
func buildImage(ctx context.Context, client *dagger.Client, platforms []dagger.Platform) ([]*dagger.Container, error) {

	fmt.Println("Docker multistage build...")

	project := client.Host().Directory(".")

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	for _, platform := range platforms {

		builderStage := client.Container().
			From(fmt.Sprintf("golang:%s-alpine", "1.21")).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", architectureOf(platform)).
			WithWorkdir("/app").
			WithDirectory(".", project, dagger.ContainerWithDirectoryOpts{
				Include: []string{"**/go.mod", "**/go.sum"},
			}).
			WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
			// run `go mod download` with only go.mod files (re-run only if mod files have changed)
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

		finalStage := client.Container(dagger.ContainerOpts{Platform: platform}).
			From("alpine:latest").
			WithLabel("org.opencontainers.image.title", "vigie").
			WithLabel("org.opencontainers.image.description", "Vigie").
			WithLabel("org.opencontainers.image.source", "https://github.com/Vincoll/vigie").
			WithLabel("org.opencontainers.image.version", Help.ShaShort).
			WithLabel("org.opencontainers.image.created", Help.DateRFC3339).
			WithFile("/vigie", builderStage.File("/app/vigie")).
			WithExec([]string{"mkdir", "--parents", "/app/config"}).
			WithEntrypoint([]string{"/vigie"}).
			WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{Args: []string{"version"}})

		platformVariants = append(platformVariants, finalStage)
	}

	return platformVariants, nil
}

func publishImage(ctx context.Context, client *dagger.Client, ctnrPlatforms []*dagger.Container, tags []string) error {

	// Publish to Registry ---
	ctr := client.Container().WithRegistryAuth("ghcr.io", "vincoll", client.SetSecret("gh_token", Sec.GITHUB_TOKEN))
	for _, tag := range tags {
		addr, err := ctr.Publish(ctx,
			fmt.Sprintf("ghcr.io/%s/vigie:%s", "vincoll", tag),
			dagger.ContainerPublishOpts{PlatformVariants: ctnrPlatforms})

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

// util that returns the architecture of the provided platform
func architectureOf(platform dagger.Platform) string {
	return platformFormat.MustParse(string(platform)).Architecture
}

/*

https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126
https://github.com/dagger/dagger/issues/4567

*/
