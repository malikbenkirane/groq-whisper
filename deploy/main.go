package main

import (
	"github.com/malikbenkirane/groq-whisper/deploy/cmd"

	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmd.NewCLI().Execute())
}
