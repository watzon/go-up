package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/watzon/go-up/internal/daemon"
	"github.com/watzon/go-up/internal/tui"
	"github.com/watzon/go-up/internal/types"
)

func initConfig() {
	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Search config in home directory with name ".go-up" (without extension)
	viper.AddConfigPath(home)
	viper.SetConfigName(".go-up")
	viper.SetConfigType("yaml")

	// Set defaults
	viper.SetDefault("daemon.host", "localhost")
	viper.SetDefault("daemon.port", 1234)

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create it with defaults
			if err := viper.SafeWriteConfig(); err != nil {
				log.Printf("Warning: Could not write default config file: %v", err)
			}
		} else {
			log.Printf("Warning: Could not read config file: %v", err)
		}
	}
}

func main() {
	var debugMode bool
	var daemonHost string
	var daemonPort int

	// Initialize config before creating commands
	initConfig()

	var rootCmd = &cobra.Command{
		Use:   "go-up",
		Short: "Starts the TUI interface",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting TUI...")
			app, err := tui.NewApp(debugMode, fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error creating TUI: %v", err)
			}
			defer app.Close() // Add this line
			app.Run()
		},
	}

	rootCmd.PersistentFlags().StringVar(&daemonHost, "host", viper.GetString("daemon.host"), "Daemon host address")
	rootCmd.PersistentFlags().IntVar(&daemonPort, "port", viper.GetInt("daemon.port"), "Daemon port")
	rootCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug mode")

	var startDaemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Starts the go-up daemon",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("Starting daemon on %s:%d...", daemonHost, daemonPort)
			daemon.Start(daemonHost, daemonPort)
		},
	}

	var addMonitorCmd = &cobra.Command{
		Use:   "add [name] [url]",
		Short: "Add a new monitor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				fmt.Println("Please provide a name and a URL to monitor")
				return
			}
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var reply string
			err = client.Call("Service.AddMonitor", struct{ Name, URL string }{args[0], args[1]}, &reply)
			if err != nil {
				log.Fatalf("Error adding monitor: %v", err)
			}
			fmt.Println(reply)
		},
	}

	var removeMonitorCmd = &cobra.Command{
		Use:   "remove [url]",
		Short: "Remove a monitor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide a URL to remove")
				return
			}
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var reply string
			err = client.Call("Service.RemoveMonitor", args[0], &reply)
			if err != nil {
				log.Fatalf("Error removing monitor: %v", err)
			}
			fmt.Println(reply)
		},
	}

	var pauseMonitorCmd = &cobra.Command{
		Use:   "pause [url]",
		Short: "Pause a monitor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide a URL to pause")
				return
			}
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var reply string
			err = client.Call("Service.PauseMonitor", args[0], &reply)
			if err != nil {
				log.Fatalf("Error pausing monitor: %v", err)
			}
			fmt.Println(reply)
		},
	}

	var resumeMonitorCmd = &cobra.Command{
		Use:   "resume [url]",
		Short: "Resume a paused monitor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide a URL to resume")
				return
			}
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var reply string
			err = client.Call("Service.ResumeMonitor", args[0], &reply)
			if err != nil {
				log.Fatalf("Error resuming monitor: %v", err)
			}
			fmt.Println(reply)
		},
	}

	var listMonitorsCmd = &cobra.Command{
		Use:   "list",
		Short: "List all monitors",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var monitors []types.Monitor
			err = client.Call("Service.ListMonitors", struct{}{}, &monitors)
			if err != nil {
				log.Fatalf("Error listing monitors: %v", err)
			}

			if len(monitors) == 0 {
				fmt.Println("No monitors found.")
			} else {
				fmt.Println("Monitors:")
				for _, monitor := range monitors {
					status := "active"
					if !monitor.IsActive {
						status = "paused"
					}
					fmt.Printf("- %s (%s) [%s]\n", monitor.Name, monitor.URL, status)
				}
			}
		},
	}

	var monitorCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Manage monitors",
	}

	var getMonitorCmd = &cobra.Command{
		Use:   "get [name]",
		Short: "Get detailed stats for a monitor",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Please provide a monitor name")
				return
			}
			client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", daemonHost, daemonPort))
			if err != nil {
				log.Fatalf("Error connecting to daemon: %v", err)
			}
			defer client.Close()

			var status types.ServiceStatus
			err = client.Call("Service.GetServiceSta`--tus", args[0], &status)
			if err != nil {
				log.Fatalf("Error getting monitor stats: %v", err)
			}

			fmt.Printf("Stats for %s (%s):\n", status.ServiceName, status.ServiceURL)
			fmt.Printf("Status: %s\n", formatStatus(status.CurrentStatus))
			fmt.Printf("Current Response Time: %dms\n", status.ResponseTime)
			fmt.Printf("Average Response Time: %.2fms\n", status.AvgResponseTime)
			fmt.Printf("Uptime (24h): %.2f%%\n", status.Uptime24Hours)
			fmt.Printf("Uptime (30d): %.2f%%\n", status.Uptime30Days)
			if !status.CertificateExpiry.IsZero() {
				fmt.Printf("Certificate Expires: %s\n", status.CertificateExpiry.Format("2006-01-02"))
			}
		},
	}

	monitorCmd.AddCommand(addMonitorCmd, removeMonitorCmd, pauseMonitorCmd, resumeMonitorCmd, listMonitorsCmd, getMonitorCmd)
	rootCmd.AddCommand(startDaemonCmd, monitorCmd)

	rootCmd.Execute()
}

func formatStatus(isUp bool) string {
	if isUp {
		return "UP"
	}
	return "DOWN"
}
