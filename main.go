package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
		if err := listAlbums(); err != nil {
			log.Fatalf("Error listing albums: %v", err)
		}
	},
}

var photosCmd = &cobra.Command{
	Use:   "photos",
	Short: "List Google Photos media items",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		if err := listPhotos(limit); err != nil {
			log.Fatalf("Error listing photos: %v", err)
		}
	},
}

var pickerCmd = &cobra.Command{
	Use:   "picker",
	Short: "Use Google Photos Picker to select photos from your entire library",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPicker(); err != nil {
			log.Fatalf("Error running picker: %v", err)
		}
	},
}

func listAlbums() error {
	config, err := getGoogleConfig()
	if err != nil {
		return fmt.Errorf("failed to get Google config: %v", err)
	}

	client := getClient(config)
	photosClient, err := NewPhotosClient(client)
	if err != nil {
		return fmt.Errorf("failed to create Photos client: %v", err)
	}

	ctx := context.Background()
	return photosClient.ListAlbums(ctx)
}

func listPhotos(limit int) error {
	config, err := getGoogleConfig()
	if err != nil {
		return fmt.Errorf("failed to get Google config: %v", err)
	}

	client := getClient(config)
	photosClient, err := NewPhotosClient(client)
	if err != nil {
		return fmt.Errorf("failed to create Photos client: %v", err)
	}

	ctx := context.Background()
	return photosClient.ListMediaItems(ctx, limit)
}

func runPicker() error {
	config, err := getGoogleConfig()
	if err != nil {
		return fmt.Errorf("failed to get Google config: %v", err)
	}

	accessToken, err := getAccessToken(config)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	client := &http.Client{}
	pickerClient := NewPickerClient(client, accessToken)
	
	ctx := context.Background()
	
	// セッションを作成
	fmt.Println("Google Photos Picker セッションを作成中...")
	session, err := pickerClient.CreateSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to create picker session: %v", err)
	}

	fmt.Printf("Google Photos Picker を開いてください:\n%s\n\n", session.PickerUri)
	fmt.Println("ブラウザで上記URLを開き、写真を選択してください...")
	
	// 選択完了を待機
	if err := pickerClient.WaitForSelection(ctx, session.Name); err != nil {
		return fmt.Errorf("failed to wait for selection: %v", err)
	}

	// 選択された写真を取得
	fmt.Println("選択された写真を取得中...")
	mediaItems, err := pickerClient.ListMediaItems(ctx, session.Name)
	if err != nil {
		return fmt.Errorf("failed to list selected media items: %v", err)
	}

	// 結果を表示
	if len(mediaItems) == 0 {
		fmt.Println("選択された写真がありません。")
		return nil
	}

	fmt.Printf("選択された写真 (%d件):\n\n", len(mediaItems))
	for i, item := range mediaItems {
		fmt.Printf("%d. %s\n", i+1, item.MediaFile.Filename)
		fmt.Printf("   ID: %s\n", item.ID)
		fmt.Printf("   Type: %s (%s)\n", item.Type, item.MediaFile.MimeType)
		fmt.Printf("   作成日時: %s\n", item.CreateTime)
		fmt.Printf("   サイズ: %dx%d\n", item.MediaFile.MediaFileMetadata.Width, item.MediaFile.MediaFileMetadata.Height)
		if item.MediaFile.MediaFileMetadata.CameraMake != "" {
			fmt.Printf("   カメラ: %s %s\n", item.MediaFile.MediaFileMetadata.CameraMake, item.MediaFile.MediaFileMetadata.CameraModel)
		}
		if item.MediaFile.MediaFileMetadata.PhotoMetadata.FocalLength > 0 {
			fmt.Printf("   撮影設定: f/%.1f, %dmm, ISO%d, %s\n", 
				item.MediaFile.MediaFileMetadata.PhotoMetadata.ApertureFNumber,
				int(item.MediaFile.MediaFileMetadata.PhotoMetadata.FocalLength),
				item.MediaFile.MediaFileMetadata.PhotoMetadata.IsoEquivalent,
				item.MediaFile.MediaFileMetadata.PhotoMetadata.ExposureTime)
		}
		fmt.Printf("   URL: %s\n", item.MediaFile.BaseUrl)
		fmt.Println()
	}

	return nil
}

func init() {
	photosCmd.Flags().IntP("limit", "l", 10, "Maximum number of photos to retrieve")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(photosCmd)
	rootCmd.AddCommand(pickerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
