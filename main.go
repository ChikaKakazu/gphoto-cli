package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gphoto-cli",
	Short: "Google Photos CLI Tool",
	Long:  "A command-line interface tool for managing Google Photos using Google API",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gphoto-cli - Google Photos CLI Tool")
		fmt.Println("Use 'gphoto-cli --help' for more information")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gphoto-cli v0.1.0")
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Google Photos albums",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing Google Photos albums...")
		fmt.Println("(Not implemented yet)")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}