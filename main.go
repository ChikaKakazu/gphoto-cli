package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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


var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup for Google OAuth credentials",
	Long:  "Configure Google OAuth 2.0 credentials through an interactive setup process",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInteractiveSetup(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View or reset configuration settings",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfigShow(); err != nil {
			log.Fatalf("Error showing config: %v", err)
		}
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration and authentication",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfigReset(); err != nil {
			log.Fatalf("Error resetting config: %v", err)
		}
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download selected photos to local directory",
	Long:  "Select photos from Google Photos and download them to a specified directory",
	Run: func(cmd *cobra.Command, args []string) {
		// 設定確認
		if !isConfigured() {
			fmt.Println("❌ Google OAuth credentials are not configured.")
			fmt.Println("Please run setup first: ./gphoto-cli setup")
			os.Exit(1)
		}

		outputDir, _ := cmd.Flags().GetString("output")
		thumbnail, _ := cmd.Flags().GetBool("thumbnail")
		
		if err := runDownloadOnly(outputDir, thumbnail); err != nil {
			log.Fatalf("Error downloading photos: %v", err)
		}
	},
}

var pickerCmd = &cobra.Command{
	Use:   "picker",
	Short: "Use Google Photos Picker to select photos from your entire library",
	Run: func(cmd *cobra.Command, args []string) {
		// 設定確認
		if !isConfigured() {
			fmt.Println("❌ Google OAuth credentials are not configured.")
			fmt.Println("Please run setup first: ./gphoto-cli setup")
			os.Exit(1)
		}

		preview, _ := cmd.Flags().GetBool("preview")
		open, _ := cmd.Flags().GetBool("open")
		download, _ := cmd.Flags().GetBool("download")
		thumbnail, _ := cmd.Flags().GetBool("thumbnail")
		
		if err := runPickerWithDisplay(preview, open, download, thumbnail); err != nil {
			log.Fatalf("Error running picker: %v", err)
		}
	},
}


func runPickerWithDisplay(preview, open, download, thumbnail bool) error {
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

	// 画像ビューアーを初期化
	var imageViewer *ImageViewer
	if preview || open || download {
		imageViewer, err = NewImageViewer(client, accessToken)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize image viewer: %v\n", err)
		} else {
			// 古い一時ファイルをクリーンアップ
			imageViewer.CleanupTempFiles()
		}
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

		// 画像表示機能
		if imageViewer != nil {
			// URLを適切に調整
			imageUrl := item.MediaFile.BaseUrl
			if thumbnail {
				imageUrl = imageViewer.GetThumbnailURL(imageUrl, 800, 600)
			} else if download || open {
				imageUrl = imageViewer.GetHighResURL(imageUrl)
			}

			// 画像をダウンロード
			if preview || open || download {
				fmt.Printf("   画像を処理中...\n")
				imagePath, err := imageViewer.DownloadImage(imageUrl, item.MediaFile.Filename)
				if err != nil {
					fmt.Printf("   Warning: Failed to download image: %v\n", err)
				} else {
					// プレビュー表示
					if preview {
						fmt.Println("   --- ASCII Preview ---")
						imageViewer.DisplayASCII(imagePath, 60)
					}

					// 外部ビューアーで開く
					if open {
						fmt.Printf("   外部ビューアーで画像を開いています...\n")
						if err := imageViewer.OpenWithDefaultViewer(imagePath); err != nil {
							fmt.Printf("   Warning: Failed to open image: %v\n", err)
						}
					}

					// ダウンロード情報を表示
					if download {
						fmt.Printf("   ダウンロード完了: %s\n", imagePath)
					}
				}
			}
		}

		fmt.Printf("   URL: %s\n", item.MediaFile.BaseUrl)
		fmt.Println()
	}

	return nil
}

func runDownloadOnly(outputDir string, thumbnail bool) error {
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

	// 出力ディレクトリの設定
	if outputDir == "" {
		outputDir = "./downloads"
	}
	
	// 出力ディレクトリを作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	fmt.Printf("📂 ダウンロード先: %s\n", outputDir)
	fmt.Printf("選択された写真 (%d件) をダウンロード中...\n\n", len(mediaItems))

	for i, item := range mediaItems {
		fmt.Printf("%d/%d: %s\n", i+1, len(mediaItems), item.MediaFile.Filename)
		
		// URLを適切に調整
		imageUrl := item.MediaFile.BaseUrl
		if thumbnail {
			imageUrl = getImageThumbnailURL(imageUrl, 800, 600)
		} else {
			imageUrl = getImageHighResURL(imageUrl)
		}

		// ファイル名を決定（元のファイル名を使用）
		filename := item.MediaFile.Filename
		if filename == "" {
			// ファイル名が空の場合はIDを使用
			ext := ".jpg" // デフォルト
			if strings.Contains(item.MediaFile.MimeType, "heif") {
				ext = ".heic"
			}
			filename = item.ID + ext
		}
		
		outputPath := filepath.Join(outputDir, filename)

		// 画像をダウンロード
		if err := downloadImageToFile(client, accessToken, imageUrl, outputPath); err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ ダウンロード完了: %s\n", outputPath)
	}

	fmt.Printf("\n🎉 すべてのダウンロードが完了しました！\n")
	fmt.Printf("📂 保存先: %s\n", outputDir)

	return nil
}

// ヘルパー関数
func getImageThumbnailURL(baseUrl string, width, height int) string {
	if strings.Contains(baseUrl, "googleusercontent.com") {
		return fmt.Sprintf("%s=w%d-h%d", baseUrl, width, height)
	}
	return baseUrl
}

func getImageHighResURL(baseUrl string) string {
	if strings.Contains(baseUrl, "googleusercontent.com") {
		return baseUrl + "=d"
	}
	return baseUrl
}

func downloadImageToFile(client *http.Client, accessToken, imageUrl, outputPath string) error {
	// 画像をダウンロード
	req, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// 認証ヘッダーを追加
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// ファイルに保存
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}

func init() {
	pickerCmd.Flags().BoolP("preview", "p", false, "Show ASCII preview in terminal")
	pickerCmd.Flags().BoolP("open", "o", false, "Open images with default viewer")
	pickerCmd.Flags().BoolP("download", "d", false, "Download images to temp directory")
	pickerCmd.Flags().Bool("thumbnail", false, "Use thumbnail size for faster download")

	downloadCmd.Flags().StringP("output", "o", "./downloads", "Output directory for downloaded images")
	downloadCmd.Flags().Bool("thumbnail", false, "Download thumbnail size instead of full resolution")

	// config サブコマンドの設定
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configResetCmd)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(pickerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
