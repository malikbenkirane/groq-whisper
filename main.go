package main

import (
	"groq/cmd"

	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmd.NewCLI(version).Execute())
}
