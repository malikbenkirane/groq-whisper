package main

import (
	"fmt"
	"os"

	"github.com/malikbenkirane/groq-whisper/cmd"

	"github.com/spf13/cobra"
)

func main() {
	cli, err := cmd.NewCLI()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cobra.CheckErr(cli.Execute())
}
