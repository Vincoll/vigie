# Vigie Makefile --------------------------------------------------------------------------

.CNTR_REGISTRY  = "vincoll"
.CNTR_REGISTRY_DEV  = "vincoll"

.GO_VERSION		= 1.21.1

.DATE           = $(shell date -u '+%Y-%m-%d_%H:%M_UTC')
.COMMIT         = $(shell git rev-parse --short HEAD)
.VIGIE_VERSION 	= $(shell ./build/scripts/vfromchangelog.sh)
.LDFLAGS    	= -ldflags "-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=$(.VIGIE_VERSION) \
							-X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=$(.DATE) \
							-X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=$(.COMMIT)"

# Protobuf
.GO_MODULE	= "github.com/vincoll/vigie"

.ROOT_DIR = $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# CONTINUOUS INTEGRATION (CI) -------------------------------------------------------------

ci-docker-all: ci-docker-clean ci-docker-mon ci-docker-backend

ci-docker-mon:
	@echo "> Create Vigie Monitoring"
	@docker network create vigie 2> /dev/null || true
	@docker compose --file build/devenv/DC_vigie_mon.yml up --detach --force-recreate --quiet-pull

ci-docker-backend:
	@echo "> Create Vigie CI Backend Containers"
	@docker compose --file build/devenv/DC_vigie_backend.yml up --detach --force-recreate --quiet-pull
	@sleep 15
	@docker exec -t VIGIE-CI_cockroach cockroach sql --file /tmp/init/db_init.sql --insecure || true
	@docker exec -t VIGIE-CI_pulsar /pulsar/bin/pulsar-admin tenants create vigie || true
	@docker exec -t VIGIE-CI_pulsar /pulsar/bin/pulsar-admin namespaces create vigie/test || true
	@docker exec -t VIGIE-CI_pulsar /pulsar/bin/pulsar-admin topics create-partitioned-topic vigie/test/test -p 1 || true
	@docker exec -t VIGIE-CI_pulsar /pulsar/bin/pulsar-admin topics create-partitioned-topic vigie/test/v0 -p 1 || true


ci-docker-clean:
	@echo "> Delete Vigie All Mon / CI Containers"
	@docker compose --file build/devenv/DC_vigie_mon.yml rm --stop --force
	@docker compose --file build/devenv/DC_vigie_backend.yml rm --stop --force

# BUILD -----------------------------------------------------------------------------------

pre-build: generate-proto

dag-build:
	dagger run go run ci/main.go

# Build the binary with your own Go env
# Output is ./bin/webapi
build-go-binary:
	GOMODULE111=on CGO_ENABLED=0 go build $(.LDFLAGS) -o ./bin/vigie
	sudo setcap cap_net_raw,cap_net_bind_service=+ep ./bin/vigie
	./bin/vigie version

# Build the binary with a Golang container
# Output is ./bin/webapi
build-go-binary-docker: test
	DOCKER_BUILDKIT=1 docker build --build-arg GO_VERSION=$(.GO_VERSION) --build-arg VIGIE_VERSION=$(.VIGIE_VERSION) --build-arg COMMIT=$(.COMMIT) --build-arg DATE=$(.DATE) \
	 			 --file "./build/release/Dockerfile.buildgobinary" --no-cache --pull \
	 			 --tag vigie:build-$(.VIGIE_VERSION) .
	@docker create -ti --name vigie_build-$(.VIGIE_VERSION) vigie:build-$(.VIGIE_VERSION) sh
	@docker cp vigie_build-$(.VIGIE_VERSION):/bin/vigie $(.ROOT_DIR)/bin
	@docker rm -f vigie_build-$(.VIGIE_VERSION)
	@docker rmi vigie:build-$(.VIGIE_VERSION)
	sudo setcap cap_net_raw,cap_net_bind_service=+ep ./bin/vigie
	./bin/vigie version

# Build Vigie docker image
# Output is a docker image webapi:$(.VIGIE_VERSION)
build-docker-image-local:
	@DOCKER_BUILDKIT=1 docker build --build-arg GO_VERSION=$(.GO_VERSION) --build-arg VIGIE_VERSION=$(.VIGIE_VERSION) --build-arg COMMIT=$(.COMMIT) --build-arg DATE=$(.DATE) \
				  --file "./Dockerfile" --no-cache --pull \
				  --tag vigie:$(.VIGIE_VERSION) .
	@docker run --tty vigie:$(.VIGIE_VERSION) version

buildx-docker-image-local:
	@DOCKER_BUILDKIT=1 docker buildx build \
					--platform=linux/arm,linux/arm64,linux/amd64 \
					--build-arg GO_VERSION=$(.GO_VERSION) --build-arg VIGIE_VERSION=$(.VIGIE_VERSION) --build-arg COMMIT=$(.COMMIT) --build-arg DATE=$(.DATE) \
				 	--file "./build/release/Dockerfile.release" --no-cache --pull \
				  	--tag vigie:$(.VIGIE_VERSION) .

# https://developers.google.com/protocol-buffers/docs/reference/go-generated
generate-proto:
	protoc --version
	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/icmp.proto
	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/tcp.proto

	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/debug.proto
	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/probe_assertion.proto
	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/probe_complete.proto
	protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/probe_metadata.proto
	#protoc --proto_path=proto/ --go_out=. --go_opt=module=github.com/vincoll/vigie proto/*.proto

#logg/logg.pb.go: logg/logg.proto
#	protoc --go_out=./ --go_opt=module=${MOD} logg/logg.proto


# PUBLISH ---------------------------------------------------------------------------------

# Goreleaser (https://goreleaser.com/)
# Build Go binaries, Create Github Packages, Create Docker Images

publish-goreleaser-dry:
	goreleaser --snapshot --skip-publish --rm-dist

publish-goreleaser-full:
	goreleaser --rm-dist

publish-docker-release-push: build-docker-image-local
	docker tag vigie:$(.VIGIE_VERSION) $(.CNTR_REGISTRY)/vigie:$(.VIGIE_VERSION)
	docker tag vigie:$(.VIGIE_VERSION) $(.CNTR_REGISTRY)/vigie:latest
	docker push $(.CNTR_REGISTRY)/vigie:$(.VIGIE_VERSION)
	docker push $(.CNTR_REGISTRY)/vigie:latest

publish-docker-dev-push: build-docker-image-local
	@docker tag vigie:$(.VIGIE_VERSION) $(.CNTR_REGISTRY)/vigie:dev
	docker push $(.CNTR_REGISTRY_DEV)/vigie:dev

publish-docker-current-push: build-docker-image-local
	docker tag vigie:$(.VIGIE_VERSION) $(.CNTR_REGISTRY)/vigie:$(.VIGIE_VERSION)
	docker push $(.CNTR_REGISTRY_DEV)/vigie:$(.VIGIE_VERSION)

# RUN -------------------------------------------------------------------------------------

run-dev: build-go-binary
	@rm -rf ./bin/test
	@cp -rf ./dev/test ./bin
	@cp -rf ./dev/var ./bin
	@(cd ./bin ; ./vigie run --config ../dev/config/vigieconf_dev.toml)

run-container-dev-demo: build-docker-image-local
	@docker run --mount type=bind,source=$(.ROOT_DIR)/dev/config/,target=/app/config/ vigie:$(.VIGIE_VERSION) run --config /app/config/vigieconf_demo_DEV.toml

run-container-prod-demo: build-docker-image-local
	@docker run --mount type=bind,source=$(.ROOT_DIR)/dev/config/,target=/app/config/ vigie:$(.VIGIE_VERSION) run --config /app/config/vigieconf_demo_PROD.toml

# DEBUG -----------------------------------------------------------------------------------

debug-docker-image:
	@docker run -ti --entrypoint sh vigie:$(.VIGIE_VERSION)

# PROFILING -------------------------------------------------------------------------------

pprof-mem:
	@go tool pprof http://127.0.0.1:6680/debug/pprof/heap

pprof-mem-inuse:
	@go tool pprof -inuse_space http://127.0.0.1:6680/debug/pprof/heap

pprof-mem-alloc:
	@go tool pprof -alloc_space http://127.0.0.1:6680/debug/pprof/heap

pprof-cpu:
	@wget http://127.0.0.1:6680/debug/pprof/profile?seconds=30

pprof-trace:
	@wget http://127.0.0.1:6680/debug/pprof/trace?seconds=5

pprof-goroutine:
	@go tool pprof http://127.0.0.1:6680/debug/pprof/goroutine?seconds=30

# TEST ------------------------------------------------------------------------------------

test:
	 go test ./... --cover

lint:
	golint -set_exit_status ./...


# DOCS ------------------------------------------------------------------------------------

# Build documentation site
docs-generate:
	make -C ./docs docs

## Serve the documentation site localy
docs-serve:
	make -C ./docs docs-serve