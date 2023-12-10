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
	service   = "vigie"
	repo      = "github.com/vincoll/vigie"
	goVersion = "1.21.5"
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
	PublishToRegistry string `env:"PUBLISH_REGISTRY,default=false,prefix=VIGIE_CI_"`
	LocalEnv          string `env:"LOCAL_ENV,default=false,prefix=VIGIE_CI_"`
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

	fmt.Println("Dagger CICD - " + service)
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(fmt.Errorf("dagger connect: %w", err))
	}

	// CI Local
	if envs.LocalEnv != "false" {
		err = CILocal(ctx, client)
		if err != nil {
			os.Exit(6)
		}
		return
	}

	// CI Pull Request

	err = CICD(ctx, client)
	if err != nil {
		os.Exit(6)
	}

}

/*

https://github.com/dagger/dagger/blob/25be91c8ea851e356563727c5a4a8c69d82f6399/internal/mage/util/util.go#L118
https://github.com/flipt-io/flipt/blob/dd47bb474870be7bb83f887a38f3b1875ebb9371/build/internal/flipt.go#L126
https://github.com/dagger/dagger/issues/4567
https://github.dev/kpenfound/greetings-api/tree/main/ci
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

		platformVariants = append(platformVariants, finalStage)
		if err != nil {
			return nil, fmt.Errorf("failed to build docker image: %w", err)
		}
		finalStage.AsService()
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

func (v *Vigie) _Serve(ctx context.Context, vigieCtnr *dagger.Container) error {

	// then in all of your tests, continue to use an explicit binding:
	pg := v.dag.Container().From("postgres:16.1-alpine").
		WithMountedDirectory("/docker-entrypoint-initdb.d/", v.dir.Directory("build/devenv/configs/sql/")).
		WithEnvVariable("POSTGRES_PASSWORD", "ci").
		WithEnvVariable("POSTGRES_USER", "ci").
		WithEnvVariable("POSTGRES_DB", "ci").
		WithExposedPort(26257).
		AsService()

	// https://docs.dagger.io/cookbook#start-and-stop-services
	vigieApiSvc := vigieCtnr.
		WithServiceBinding("pg", pg).
		WithExposedPort(6680).
		WithExposedPort(6690).
		WithMountedFile("/app/config/vigie_ci.toml", v.dir.File("build/ci/configs/vigie/vigieconf_api.toml")).
		WithEntrypoint([]string{"/vigie"}).
		WithExec([]string{"api", "--config", "/app/config/vigie_api.toml"}).
		AsService()

	// expose web service to host
	tunnel, err := v.dag.Host().Tunnel(vigieApiSvc).Start(ctx)
	if err != nil {
		panic(err)
	}
	defer tunnel.Stop(ctx)

	// get web service address
	srvAddr, err := tunnel.Endpoint(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Vigie API is running at: %s\n", srvAddr)
	return nil
}

func (v *Vigie) Serve(ctx context.Context, vigieCtnr *dagger.Container) error {

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
	vc := v.dag.LoadContainerFromID(img).
		//		WithServiceBinding("docker", dockerd).
		WithServiceBinding("pg", pg).
		WithExposedPort(6680). // API
		WithExposedPort(6690). // Tech (metrics, health, pprof)
		WithMountedDirectory("/app/config/", v.dir.Directory("build/ci/configs/vigie/")).
		WithEntrypoint([]string{"/vigie"}).
		WithExec([]string{"api", "--config", "/app/config/vigieconf_ci.toml"}).
		AsService()
	if err != nil {
		return err
	}

	fmt.Sprint(vc)

	return nil
}

func (v *Vigie) IntegrationTest(ctx context.Context, ctnr *dagger.Container) error {

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

func CICD(ctx context.Context, client *dagger.Client) error {

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
	ctnrs, err := vigieCI.BuildImage(ctx, goVersion, targetArch)
	if err != nil {
		panic(fmt.Errorf("build docker image: %w", err))
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

	err = vigieCI.Serve(ctx, ctnr[0])

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
