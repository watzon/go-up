package main

import (
    "github.com/spf13/cobra"
    "log"
    "github.com/watzon/go-up/internal/daemon"
    "github.com/watzon/go-up/internal/tui"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "go-up",
        Short: "Starts the TUI interface",
        Run: func(cmd *cobra.Command, args []string) {
            log.Println("Starting TUI...")
            tui.Start()
        },
    }

    var startDaemonCmd = &cobra.Command{
        Use:   "daemon",
        Short: "Starts the go-up daemon",
        Run: func(cmd *cobra.Command, args []string) {
            log.Println("Starting daemon...")
            daemon.Start()
        },
    }

    rootCmd.AddCommand(startDaemonCmd)
    rootCmd.Execute()
}
