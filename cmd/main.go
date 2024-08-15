package main

import (
	"github.com/golang/glog"
	"github.com/morvencao/event-based-transport-demo/cmd/agent"
	"github.com/morvencao/event-based-transport-demo/cmd/source"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "event-based-transport-demo",
		Long: "Event Based Transport Demo",
	}

	// All subcommands under root
	sourceCmd := source.NewSourceCommand()
	agentCmd := agent.NewAgentCommand()

	// Add subcommand(s)
	rootCmd.AddCommand(sourceCmd, agentCmd)

	if err := rootCmd.Execute(); err != nil {
		glog.Fatalf("error running command: %v", err)
	}
}
