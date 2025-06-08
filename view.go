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

	fmt.Println("🖼️  Quick View Mode - Select photos and view metadata")
	return runPicker() // 画像表示機能を削除し、基本的なpicker機能のみ使用
}

func init() {
	rootCmd.AddCommand(viewCmd)
}