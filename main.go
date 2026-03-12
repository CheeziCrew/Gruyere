package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/CheeziCrew/gruyere/cli"
	"github.com/CheeziCrew/gruyere/tui"
)

func main() {
	if len(os.Args) > 1 {
		root := cli.BuildCLI()
		if err := root.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(tui.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
