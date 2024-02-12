package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
	"github.com/sethvargo/go-envconfig"

	"main/api"
	"main/build"
	"main/common"
)

// Warn: I'm experimenting with dagger.io
// Don't use this code bindly

const (
	service       = "vigie"
	repo          = "github.com/vincoll/vigie"
	goVersion     = "1.22.0" // https://hub.docker.com/_/golang
	alpineVersion = "3.19"   // https://hub.docker.com/_/alpine
)

var (
	targetArch = []dagger.Platform{"linux/amd64", "linux/arm64"}
	imageTags  = []string{"latest", buildCtx.ShaShort}
	// Build Context
	buildCtx = common.NewBuildContext()
)

var envs Environments // Global
type Environments struct {
	PublishToRegistry string `env:"PUBLISH_REGISTRY,default=false"`
}

var Secs Secrets // Global
type Secrets struct {
	GITHUB_TOKEN string `env:"GITHUB_TOKEN,default="`
}

// init() is invoked before main()
// It will initialize the vars based on the build context
func init() {

	// Env Var
	if err := envconfig.Process(context.Background(), &envs); err != nil {
		log.Fatal(err)
	}

	// Secrets
	if err := envconfig.Process(context.Background(), &Secs); err != nil {
		log.Fatal(err)
	}

}

// main is the entrypoint of the CI
// It will run the CI based on the build context digested by init()
func main() {

	ctx := context.Background()
	defer dag.Close()

	// Access all arguments as a slice using os.Args
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Missing arguments: go run ci/main.go <CICD_MODE>")
		fmt.Println("Example: CICD_MODE [local, pr, release]")
		os.Exit(1)
	}

	CICDMode := args[1]
	fmt.Printf("Dagger CICD - %s : %s\n", service, CICDMode)

	switch CICDMode {

	// CI Pull Request
	case "pr", "pullrequest":
		err := CIPullRequest(ctx)
		if err != nil {
			os.Exit(8)
		}

	// CI Release
	case "release":
		err := CICDRelease(ctx)
		if err != nil {
			os.Exit(9)
		}

	case "local":
		// CI Local
		err := CILocal(ctx)
		if err != nil {
			os.Exit(3)
		}

	default:
		fmt.Printf("CICD_MODE %q is not supported\n", CICDMode)
		os.Exit(2)
	}

}

/*
https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126
https://github.com/dagger/dagger/issues/4567
https://github.dev/kpenfound/greetings-api/tree/main/ci
https://github.com/portward/portward/pull/45/files#diff-e99d527b8183955c3241c07f61a239d20440e5a6aab4fa41223c0e7292814709
*/

type Vigie struct {
	dir *dagger.Directory
	vb  *build.VigieBuild
}

func newVigie(ctx context.Context) *Vigie {

	var v Vigie

	v.dir = dag.Host().Directory(".",
		dagger.HostDirectoryOpts{
			Exclude: []string{".git", ".vscode", "docs"},
		})
	v.vb = build.NewVigieBuild(v.dir, &buildCtx)
	return &v
}

// BuildImage builds the docker (multi-arch) images for the provided platforms
func (v *Vigie) BuildImage(ctx context.Context, goVer string, platforms []dagger.Platform) ([]*dagger.Container, error) {

	containers, err := v.vb.BuildImage(ctx, goVer, alpineVersion, platforms)
	if err != nil {
		return nil, fmt.Errorf("build docker image: %w", err)
	}
	return containers, nil
}

// PublishImage publishes the docker (multi-arch) images to a registry
func (v *Vigie) PublishImage(ctx context.Context, ctnrPlatforms []*dagger.Container, tags []string) error {

	if Secs.GITHUB_TOKEN == "" {
		return fmt.Errorf("env Var GITHUB_TOKEN is not set. Tips: export GITHUB_TOKEN=$(gh auth token)")
	}
	imageTags2 := []string{"latest", buildCtx.ShaShort}

	fmt.Printf("Publishing Image to: %s\n", imageTags2)
	// Publish to Registry ---
	ctr := dag.Container().WithRegistryAuth("ghcr.io", "vincoll", dag.SetSecret("gh_token", Secs.GITHUB_TOKEN))
	for _, tag := range imageTags2 {
		fullImageTag := fmt.Sprintf("ghcr.io/%s/%s:%s", "vincoll", service, tag)
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

// Runs integration tests on different part of the app
// Takes the container built during this CI
func (v *Vigie) IntegrationTest(ctx context.Context, vigieCtnr *dagger.Container) error {

	// Vigie API
	vigieApi := api.NewVigieAPI(v.dir, vigieCtnr)

	err := vigieApi.IntegrationTest(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (v *Vigie) UnitTest(ctx context.Context) error {

	return nil
}

func (v *Vigie) Lint(ctx context.Context) error {

	return nil

	// https://github.com/golangci/golangci-lint

	_, err := dag.Container().
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

// CI in release context
func CICDRelease(ctx context.Context) error {

	vigieCI := newVigie(ctx)

	//
	// Lint
	//
	err := vigieCI.Lint(ctx)
	if err != nil {
		return fmt.Errorf("lint with golangci-lint : %w", err)
	}

	//
	// Test
	//
	err = vigieCI.UnitTest(ctx)
	if err != nil {
		panic(fmt.Errorf("unit test failed: %w", err))
	}

	//
	// Docker build on current arch
	//
	ctnrs, err := vigieCI.vb.BuildImage(ctx, goVersion, alpineVersion, targetArch)
	if err != nil {
		return fmt.Errorf("build docker image: %w", err)
	}

	//
	// Test
	//
	fmt.Println("INTEGRATION TEST")
	err = vigieCI.IntegrationTest(ctx, ctnrs[0])
	if err != nil {
		panic(fmt.Errorf("test failed: %w", err))
	}

	//
	// Publish
	//
	fmt.Println("PUBLISH")
	err = vigieCI.PublishImage(ctx, ctnrs, imageTags)
	if err != nil {
		panic(fmt.Errorf("publish docker image: %w", err))
	}

	return nil
}

// CI in PR context
func CIPullRequest(ctx context.Context) error {

	vigieCI := newVigie(ctx)

	//
	// Lint
	//
	err := vigieCI.Lint(ctx)
	if err != nil {
		return fmt.Errorf("lint with golangci-lint : %w", err)
	}

	//
	// Docker build on current arch
	//
	curTargetArch := []dagger.Platform{dagger.Platform(fmt.Sprintf("linux/%s", runtime.GOARCH))}
	ctnr, err := vigieCI.vb.BuildImage(ctx, goVersion, alpineVersion, curTargetArch)
	if err != nil {
		return fmt.Errorf("build docker image: %w", err)
	}

	//
	// Test
	//
	err = vigieCI.IntegrationTest(ctx, ctnr[0])
	if err != nil {
		panic(fmt.Errorf("test failed: %w", err))
	}

	fmt.Println("Done")

	return nil
}

// CI in local context
// Runs CI localy, on the current arch
// Does not publish, nor deploy
func CILocal(ctx context.Context) error {

	fmt.Println("LOCAL Env")

	vigieCI := newVigie(ctx)

	//
	// Lint
	//
	err := vigieCI.Lint(ctx)
	if err != nil {
		return fmt.Errorf("lint with golangci-lint : %w", err)
	}

	//
	// Test
	//
	err = vigieCI.UnitTest(ctx)
	if err != nil {
		panic(fmt.Errorf("unit test failed: %w", err))
	}

	//
	// Docker build on current arch
	//
	curTargetArch := []dagger.Platform{dagger.Platform(fmt.Sprintf("linux/%s", runtime.GOARCH))}
	ctnr, err := vigieCI.vb.BuildImage(ctx, goVersion, alpineVersion, curTargetArch)
	if err != nil {
		return fmt.Errorf("build docker image: %w", err)
	}
	
	//
	// Test
	//
	fmt.Println("INTEGRATION TEST")
	err = vigieCI.IntegrationTest(ctx, ctnr[0])
	if err != nil {
		panic(fmt.Errorf("test failed: %w", err))
	}

	fmt.Println("Done")

	return nil
}

func CIremote(sha string) error {

	return nil
}
