package main

import (
	"os"

	"github.com/lum-network/chain/cmd/chaind/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := cmd.Execute(rootCmd); err != nil {
		os.Exit(1)
	}
}
