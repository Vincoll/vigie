package api

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
)

type VigieAPI struct {
	dir  *dagger.Directory
	ctnr *dagger.Container
}

func NewVigieAPI(dir *dagger.Directory, ctnr *dagger.Container) *VigieAPI {
	return &VigieAPI{
		dir:  dir,
		ctnr: ctnr,
	}
}

func (v *VigieAPI) IntegrationTest(ctx context.Context) error {

	vigieApi, err := v.serveAPI(ctx)

	// https://docs.usebruno.com/
	_, err = dag.Container().
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
		return fmt.Errorf("Vigie API Integration Test failed: %s", err)
	}

	return nil
}

func (v *VigieAPI) serveAPI(ctx context.Context) (*dagger.Service, error) {

	//	dockerd, _ := v.dag.Container().From("docker:dind").AsService().Start(ctx)

	pg := dag.Container().
		From("postgres:16.1-alpine").
		//		WithServiceBinding("docker", dockerd).
		WithMountedDirectory("/docker-entrypoint-initdb.d/", v.dir.Directory("/build/devenv/configs/sql/")).
		WithEnvVariable("POSTGRES_PASSWORD", "ci").
		WithEnvVariable("POSTGRES_USER", "ci").
		WithEnvVariable("POSTGRES_DB", "ci").
		WithExposedPort(5432).
		AsService()

	img, err := v.ctnr.ID(ctx)
	vigieApi := dag.LoadContainerFromID(img).
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
