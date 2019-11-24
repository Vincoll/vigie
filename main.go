//go:generate statik -src=./webui

package main

import "github.com/vincoll/vigie/cmd/vigie"

func main() {
	vigiemain.Main()
}
