package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Quick view mode - select and immediately view photos",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runQuickView(); err != nil {
			log.Fatalf("Error in view mode: %v", err)
		}
	},
}

func runQuickView() error {
	// 設定確認
	if !isConfigured() {
		fmt.Println("❌ Google OAuth credentials are not configured.")
		fmt.Println("Please run setup first: ./gphoto-cli setup")
		return fmt.Errorf("not configured")
	}

	fmt.Println("🖼️  Quick View Mode - Select photos to view immediately")
	return runPickerWithDisplay(true, true, false, true) // preview=true, open=true, download=false, thumbnail=true
}

func init() {
	rootCmd.AddCommand(viewCmd)
}