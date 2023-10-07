package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	contextDir := client.Host().Directory(".")

	ref, err := contextDir.
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "Dockerfile",

			BuildArgs: []dagger.BuildArg{
				{Name: "GO_VERSION", Value: "1.21"},
				{Name: "DATE", Value: "1.21"}
				{Name: "COMMIT", Value: "1.21"}},
			Target: "final",
		}).Stdout(ctx)
	//Publish(ctx, fmt.Sprintf("ttl.sh/hello-dagger-%.0f", math.Floor(rand.Float64()*10000000))) //#nosec
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to :%s\n", ref)
}
