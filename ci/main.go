package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"dagger.io/dagger"
	platformFormat "github.com/containerd/containerd/platforms"
	"github.com/sethvargo/go-envconfig"
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
	imageTags  = []string{"latest", vars.ShaShort}
)

var vars Vars // Global
type Vars struct {
	DateRFC3339 string
	Sha         string
	ShaShort    string
	Version     string
}

var envs Environements // Global
type Environements struct {
	PublishToRegistry string `env:"PUBLISH_REGISTRY,default=false"`
	CICDMode          string `env:"CICD_MODE,default=local"`
}

var Secs Secrets // Global
type Secrets struct {
	GITHUB_TOKEN string `env:"GITHUB_TOKEN,default="`
}

// init() is invoked before main()
// It will initialize the vars based on the build context
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

// main is the entrypoint of the CI
// It will run the CI based on the build context digested by init()
func main() {

	ctx := context.Background()

	fmt.Printf("Dagger CICD - %s : %s\n", service, envs.CICDMode)
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(fmt.Errorf("dagger connect: %w", err))
	}

	switch envs.CICDMode {

	// CI Pull Request
	case "pr", "pullrequest":
		err = CIPullRequest(ctx, client)
		if err != nil {
			os.Exit(8)
		}

	// CI Release
	case "release":
		err = CICDRelease(ctx, client)
		if err != nil {
			os.Exit(9)
		}

	case "local":
		// CI Local
		err = CILocal(ctx, client)
		if err != nil {
			os.Exit(3)
		}

	default:
		fmt.Printf("CICD_MODE %q is not supported\n", envs.CICDMode)
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
	dag *dagger.Client
	dir *dagger.Directory
}

func newVigie(ctx context.Context, d *dagger.Client) *Vigie {

	var v Vigie

	v.dag = d
	v.dir = d.Host().Directory(".",
		dagger.HostDirectoryOpts{
			Exclude: []string{".git", "docs"},
		})

	return &v
}

func (v *Vigie) BuildImage(ctx context.Context, goVer string, platforms []dagger.Platform) ([]*dagger.Container, error) {

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	fmt.Printf("Building OCI Images for platforms: %v\n", platforms)
	for _, platform := range platforms {
		fmt.Printf("Building on %v ... \n", platform)

		// Build the binary
		builderStage := v.dag.Container().
			From(fmt.Sprintf("golang:%s-alpine%s", goVer, alpineVersion)).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", architectureOf(platform)).
			WithWorkdir("/app").
			WithDirectory(".", v.dir, dagger.ContainerWithDirectoryOpts{
				Include: []string{"**/go.mod", "**/go.sum"},
			}).
			// include a cache for go build
			WithMountedCache("/go/pkg/mod", v.dag.CacheVolume("go-mod")).
			WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
			WithMountedCache("/go/build-cache", v.dag.CacheVolume("go-build")).
			WithEnvVariable("GOCACHE", "/go/build-cache").

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
		finalStage, err := v.dag.Container(dagger.ContainerOpts{Platform: platform}).
			From("alpine:"+alpineVersion).
			WithLabel("org.opencontainers.image.title", service).
			WithLabel("org.opencontainers.image.description", "Vigie").
			WithLabel("org.opencontainers.image.source", "https://github.com/Vincoll/vigie").
			WithLabel("org.opencontainers.image.version", vars.ShaShort).
			WithLabel("org.opencontainers.image.created", vars.DateRFC3339).
			WithFile("/vigie", builderStage.File("/app/vigie")).
			WithExec([]string{"mkdir", "--parents", "/app/config"}).
			WithEntrypoint([]string{"/vigie"}).
			WithDefaultArgs([]string{"version"}).Sync(ctx)

		platformVariants = append(platformVariants, finalStage)
		if err != nil {
			return nil, fmt.Errorf("failed to build docker image: %w", err)
		}
		finalStage.AsService()
	}

	return platformVariants, nil
}

func (v *Vigie) PublishImage(ctx context.Context, ctnrPlatforms []*dagger.Container, tags []string) error {

	if Secs.GITHUB_TOKEN == "" {
		return fmt.Errorf("env Var GITHUB_TOKEN is not set. Tips: export GITHUB_TOKEN=$(gh auth token)")
	}
	imageTags2 := []string{"latest", vars.ShaShort}

	fmt.Printf("Publishing Image to: %s\n", imageTags2)
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

func (v *Vigie) serveAPI(ctx context.Context, vigieCtnr *dagger.Container) (*dagger.Service, error) {

	//	dockerd, _ := v.dag.Container().From("docker:dind").AsService().Start(ctx)

	pg := v.dag.Container().
		From("postgres:16.1-alpine").
		//		WithServiceBinding("docker", dockerd).
		WithMountedDirectory("/docker-entrypoint-initdb.d/", v.dir.Directory("/build/devenv/configs/sql/")).
		WithEnvVariable("POSTGRES_PASSWORD", "ci").
		WithEnvVariable("POSTGRES_USER", "ci").
		WithEnvVariable("POSTGRES_DB", "ci").
		WithExposedPort(5432).
		AsService()

	img, err := vigieCtnr.ID(ctx)
	vigieApi := v.dag.LoadContainerFromID(img).
		//		WithServiceBinding("docker", dockerd).
		WithServiceBinding("pg", pg).
		WithExposedPort(6680). // API
		WithExposedPort(6690). // Tech (metrics, health, pprof)
		WithMountedDirectory("/app/config/", v.dir.Directory("build/ci/configs/vigie/")).
		WithEntrypoint([]string{"/vigie"}).
		WithExec([]string{"api", "--config", "/app/config/vigieconf_ci.toml"}).
		AsService()
	if err != nil {
		return nil, err
	}

	return vigieApi, nil
}

func (v *Vigie) IntegrationTest(ctx context.Context, vigieCtnr *dagger.Container) error {

	vigieApi, err := v.serveAPI(ctx, vigieCtnr)

	// https://docs.usebruno.com/
	_, err = v.dag.Container().
		From("vincoll/bruno:latest").
		//		WithServiceBinding("docker", dockerd).
		WithServiceBinding("vigie-api", vigieApi).
		WithEnvVariable("VIGIE_API_FQDN", "vigie-api").
		WithMountedDirectory("/tmp/", v.dir.Directory("build/tests/api/Vigie")).
		WithWorkdir("/tmp/").
		WithEntrypoint([]string{"bru"}).
		WithExec([]string{"run", "api", "-r", "--env", "ci"}).
		Stdout(ctx)
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

func (v *Vigie) ComposeUp(ctx context.Context) error {

	return nil
}

func CICDRelease(ctx context.Context, client *dagger.Client) error {

	defer client.Close()

	vigieCI := newVigie(ctx, client)

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
	ctnrs, err := vigieCI.BuildImage(ctx, goVersion, targetArch)
	if err != nil {
		panic(fmt.Errorf("build docker image: %w", err))
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
	err = vigieCI.PublishImage(ctx, ctnrs, imageTags)
	if err != nil {
		panic(fmt.Errorf("publish docker image: %w", err))
	}

	return nil
}

// CI in PR context
func CIPullRequest(ctx context.Context, client *dagger.Client) error {

	defer client.Close()

	vigieCI := newVigie(ctx, client)

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
	ctnr, err := vigieCI.BuildImage(ctx, goVersion, curTargetArch)
	if err != nil {
		panic(fmt.Errorf("build docker image: %w", err))
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
func CILocal(ctx context.Context, client *dagger.Client) error {

	fmt.Println("LOCAL Env")
	defer client.Close()

	vigieCI := newVigie(ctx, client)

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
	ctnr, err := vigieCI.BuildImage(ctx, goVersion, curTargetArch)
	if err != nil {
		panic(fmt.Errorf("build docker image: %w", err))
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

//
// Tools
//

// architectureOf is a util that returns the architecture of the provided platform
func architectureOf(platform dagger.Platform) string {
	return platformFormat.MustParse(string(platform)).Architecture
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
