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
	"github.com/sethvargo/go-envconfig"
)

// Warn: I'm experimenting with dagger.io
// Don't use this code bindly

const (
	service   = "vigie"
	goVersion = "1.21"
)

var (
	targetArch = []dagger.Platform{"linux/amd64", "linux/arm64"}
	imageTags  = []string{"latest", vars.ShaShort}
)

var vars Vars
type Vars struct {
	DateRFC3339 string
	Sha         string
	ShaShort    string
	Version     string
}

var envs Environements
type Environements struct {
	PublishToRegistry string `env:"PUBLISH_REGISTRY,default=false,prefix=VIGIE_CI_"`
}

var Secs Secrets
type Secrets struct {
	GITHUB_TOKEN string `env:"GITHUB_TOKEN,default=notSet"`
}

func init() {

	// Vars
	vars.ShaShort, vars.Sha = getSHA()
	vars.DateRFC3339 = time.Now().Format(time.RFC3339)
	vars.Version = "0.0.1"

	// Env
	if err := envconfig.Process(context.Background(), &envs); err != nil {
		log.Fatal(err)
	}

	// Secrets
	if err := envconfig.Process(context.Background(), &Secs); err != nil {
		log.Fatal(err)
	}

}

func main() {
	ctx := context.Background()

	fmt.Println("Dagger CICD - " + service)

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	vigieCI := NewVigie(ctx, client)

	//
	// Lint
	//
	err = vigieCI.Lint(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to lint with golangci-lint : %w", err))
	}

	//
	// Docker multistage build
	//
	err = buildAndPublishImage(ctx, client, targetArch, imageTags, false)
	if err != nil {
		panic(fmt.Errorf("failed to build docker image: %w", err))
	}

	fmt.Println("Done")
}

/*

https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126
https://github.com/dagger/dagger/issues/4567

*/

/////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////

// buildAndPublishImage builds a multiArch docker image and publishes it to a registry is enabled
func buildAndPublishImage(ctx context.Context, client *dagger.Client, archs []dagger.Platform, tags []string, pushImage bool) error {

	cntrs, err := buildImage(ctx, client, archs)
	if err != nil {
		panic(fmt.Errorf("failed to build docker image: %w", err))
	}

	if envs.PublishToRegistry != "false" {
		err = publishImage(ctx, client, cntrs, tags)
		if err != nil {
			panic(fmt.Errorf("failed to publish image: %w", err))
		}
	}
	return nil
}

// buildImage builds a multiArch docker image with a multistage build
func buildImage(ctx context.Context, client *dagger.Client, platforms []dagger.Platform) ([]*dagger.Container, error) {

	fmt.Println("Docker multistage build...")

	project := client.Host().Directory(".", dagger.HostDirectoryOpts{Exclude: []string{".git"}})

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	fmt.Printf("Building OCI Images for platforms: %v\n", platforms)
	for _, platform := range platforms {
		fmt.Printf("Building: %v ... ", platform)

		builderStage := client.Container().
			From(fmt.Sprintf("golang:%s-alpine", "1.21")).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", architectureOf(platform)).
			WithWorkdir("/app").
			WithDirectory(".", project, dagger.ContainerWithDirectoryOpts{
				Include: []string{"**/go.mod", "**/go.sum"},
			}).
			// include a cache for go build
			WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod")).
			WithMountedCache("/root/.cache/go-build", client.CacheVolume("go-build")).

			// run `go mod download` with only go.mod files (re-run only if mod files have changed)
			WithExec([]string{"go", "mod", "download"}).

			// run `go build` with all source
			WithMountedDirectory(".", project).
			WithExec([]string{"go", "build",
				"-ldflags",
				"-X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=" + vars.ShaShort + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=" + vars.DateRFC3339 + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=" + vars.Version + " ",
				"-o", "vigie"})

		finalStage, _ := client.Container(dagger.ContainerOpts{Platform: platform}).
			From("alpine:latest").
			WithLabel("org.opencontainers.image.title", service).
			WithLabel("org.opencontainers.image.description", "Vigie").
			WithLabel("org.opencontainers.image.source", "https://github.com/Vincoll/vigie").
			WithLabel("org.opencontainers.image.version", vars.ShaShort).
			WithLabel("org.opencontainers.image.created", vars.DateRFC3339).
			WithFile("/vigie", builderStage.File("/app/vigie")).
			WithExec([]string{"mkdir", "--parents", "/app/config"}).
			WithEntrypoint([]string{"/vigie"}).
			WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{Args: []string{"version"}}).Sync(ctx)

		fmt.Println("DONE")

		platformVariants = append(platformVariants, finalStage)
	}

	return platformVariants, nil
}

// publishImage publishes a multiArch docker image to a registry
func publishImage(ctx context.Context, client *dagger.Client, ctnrPlatforms []*dagger.Container, tags []string) error {

	if Secs.GITHUB_TOKEN == "notSet" {
		return fmt.Errorf("env Var GITHUB_TOKEN is not set. Tips: export GITHUB_TOKEN=$(gh auth token)")
	}
	imageTags2 := []string{"latest", vars.ShaShort}
	fmt.Printf("Publishing Image to: %s", imageTags2)
	// Publish to Registry ---
	ctr := client.Container().WithRegistryAuth("ghcr.io", "vincoll", client.SetSecret("gh_token", Secs.GITHUB_TOKEN))
	for _, tag := range imageTags2 {
		fullImageTag := fmt.Sprintf("ghcr.io/%s/vigie:%s", "vincoll", tag)
		fmt.Printf("Publishing Image to: %s", fullImageTag)
		addr, err := ctr.Publish(ctx,
			fullImageTag,
			dagger.ContainerPublishOpts{PlatformVariants: ctnrPlatforms})

		if err != nil {
			return err
		}
		fmt.Printf("published image to :%s\n", addr)
	}

	return nil
}

// architectureOf is a util that returns the architecture of the provided platform
func architectureOf(platform dagger.Platform) string {
	return platformFormat.MustParse(string(platform)).Architecture
}

/////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////



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

/////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////

const (
	APP   = "vigie"
	REPO  = "github.com/vincoll/vigie"
	IMAGE = ""
)

type Vigie struct {
	dag *dagger.Client
	dir *dagger.Directory
}

func NewVigie(ctx context.Context, d *dagger.Client) *Vigie {

	var v Vigie

	v.dag = d
	v.dir = d.Host().Directory(".",
		dagger.HostDirectoryOpts{
			Include: []string{"*.*"},
			Exclude: []string{".git", "docs"},
		})

	return &v
}

func (v *Vigie) BuildImage(ctx context.Context, goVer string, platforms []dagger.Platform) ([]*dagger.Container, error) {

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	fmt.Printf("Building OCI Images for platforms: %v\n", platforms)
	for _, platform := range platforms {
		fmt.Printf("Building on %v ... ", platform)

		// Build the binary
		builderStage := v.dag.Container().
			From(fmt.Sprintf("golang:%s-alpine", goVer)).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", architectureOf(platform)).
			WithWorkdir("/app").
			WithDirectory(".", v.dir, dagger.ContainerWithDirectoryOpts{
				Include: []string{"**/go.mod", "**/go.sum"},
			}).
			// include a cache for go build
			WithMountedCache("/go/pkg/mod", v.dag.CacheVolume("go-mod")).
			WithMountedCache("/root/.cache/go-build", v.dag.CacheVolume("go-build")).

			// run `go mod download` with only go.mod files (re-run only if mod files have changed)
			WithExec([]string{"go", "mod", "download"}).

			// run `go build` with all source
			WithMountedDirectory(".", v.dir).
			WithExec([]string{"go", "build",
				"-ldflags",
				"-X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=" + vars.ShaShort + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=" + vars.DateRFC3339 + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=" + vars.Version + " ",
				"-o", "vigie"})

		// Extract the binary from the builder stage and create the final stage
		finalStage, _ := v.dag.Container(dagger.ContainerOpts{Platform: platform}).
			From("alpine:latest").
			WithLabel("org.opencontainers.image.title", service).
			WithLabel("org.opencontainers.image.description", "Vigie").
			WithLabel("org.opencontainers.image.source", "https://github.com/Vincoll/vigie").
			WithLabel("org.opencontainers.image.version", vars.ShaShort).
			WithLabel("org.opencontainers.image.created", vars.DateRFC3339).
			WithFile("/vigie", builderStage.File("/app/vigie")).
			WithExec([]string{"mkdir", "--parents", "/app/config"}).
			WithEntrypoint([]string{"/vigie"}).
			WithDefaultArgs(dagger.ContainerWithDefaultArgsOpts{Args: []string{"version"}}).Sync(ctx)

		fmt.Println("DONE")

		platformVariants = append(platformVariants, finalStage)
	}

	return platformVariants, nil
}

func (v *Vigie) PublishImage(ctx context.Context, ctnrPlatforms []*dagger.Container, tags []string) error {

	if Secs.GITHUB_TOKEN == "notSet" {
		return fmt.Errorf("env Var GITHUB_TOKEN is not set. Tips: export GITHUB_TOKEN=$(gh auth token)")
	}
	imageTags2 := []string{"latest", vars.ShaShort}

	fmt.Printf("Publishing Image to: %s", imageTags2)
	// Publish to Registry ---
	ctr := v.dag.Container().WithRegistryAuth("ghcr.io", "vincoll", v.dag.SetSecret("gh_token", Secs.GITHUB_TOKEN))
	for _, tag := range imageTags2 {
		fullImageTag := fmt.Sprintf("ghcr.io/%s/vigie:%s", "vincoll", tag)
		fmt.Printf("Publishing Image to: %s", fullImageTag)
		addr, err := ctr.Publish(ctx,
			fullImageTag,
			dagger.ContainerPublishOpts{PlatformVariants: ctnrPlatforms})

		if err != nil {
			return err
		}
		fmt.Printf("published image to :%s\n", addr)
	}

	return nil
}

func (v *Vigie) Serve(ctx context.Context) error {

	return nil
}

func (v *Vigie) Test(ctx context.Context) error {

	return nil
}

func (v *Vigie) Lint(ctx context.Context) error {

	return nil

	// https://github.com/golangci/golangci-lint

	_, err := v.dag.Container().
		From("golangci/golangci-lint:v1.54-alpine").
		WithMountedDirectory("/app", v.dir).
		WithWorkdir("/app").
		WithEnvVariable("GOGC", "100"). // Default value : 100 https://golangci-lint.run/usage/performance/
		WithExec([]string{"golangci-lint", "run", "-v", "--timeout", "5m"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GoCI() error {

}
