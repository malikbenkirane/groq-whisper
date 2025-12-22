package main

import (
	"github.com/malikbenkirane/groq-whisper/setup/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cobra.CheckErr(cmd.NewCLI().Execute())
}
