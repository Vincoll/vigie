package build

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
	platformFormat "github.com/containerd/containerd/platforms"

	"main/common"
)

type VigieBuild struct {
	dir           *dagger.Directory
	buildCtx      *common.BuildContext

}

func NewVigieBuild(dir *dagger.Directory, buildCtx *common.BuildContext) *VigieBuild {
	return &VigieBuild{
		dir:          dir,
		buildCtx:     buildCtx,
	}
}

// BuildImage builds the docker (multi-arch) images for the provided platforms
func (vb *VigieBuild) BuildImage(ctx context.Context, goVer string,alpineVersion string ,platforms []dagger.Platform) ([]*dagger.Container, error) {

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	fmt.Printf("Building OCI Images for platforms: %v\n", platforms)
	for _, platform := range platforms {
		fmt.Printf("Building on %v ... \n", platform)

		// Get the architecture from
		ctnrPlatformArch := platformFormat.MustParse(string(platform)).Architecture

		// Build the binary
		builderStage := dag.Container().
			From(fmt.Sprintf("golang:%s-alpine%s", goVer, alpineVersion)).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", "linux").
			WithEnvVariable("GOARCH", ctnrPlatformArch).
			WithWorkdir("/app").
			WithDirectory(".", vb.dir, dagger.ContainerWithDirectoryOpts{
				Include: []string{"**/go.mod", "**/go.sum"},
			}).
			// include a cache for go build
			WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
			WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
			WithMountedCache("/go/build-cache", dag.CacheVolume("go-build")).
			WithEnvVariable("GOCACHE", "/go/build-cache").

			// run `go mod download` with only go.mod files (re-run only if mod files have changed)
			WithExec([]string{"go", "mod", "download"}).

			// run `go build` with all source
			WithMountedDirectory(".", vb.dir).
			WithExec([]string{"go", "build",
				"-ldflags",
				"-X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=" + vb.buildCtx.ShaShort + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=" + vb.buildCtx.DateRFC3339 + " " +
					"-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=" + vb.buildCtx.Version + " ",
				"-o", "vigie"})

		// Extract the binary from the builder stage and create the final stage
		finalStage, err := dag.Container(dagger.ContainerOpts{Platform: platform}).
			From("alpine:"+alpineVersion).
			WithLabel("org.opencontainers.image.title", "Vigie").
			WithLabel("org.opencontainers.image.description", "Vigie").
			WithLabel("org.opencontainers.image.source", "https://github.com/Vincoll/vigie").
			WithLabel("org.opencontainers.image.version",vb.buildCtx.ShaShort).
			WithLabel("org.opencontainers.image.created", vb.buildCtx.DateRFC3339).
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
