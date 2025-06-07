package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

type ImageViewer struct {
	httpClient  *http.Client
	accessToken string
	tempDir     string
}

func NewImageViewer(httpClient *http.Client, accessToken string) (*ImageViewer, error) {
	// 一時ディレクトリを作成
	tempDir := filepath.Join(os.TempDir(), "gphoto-cli")
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}

	return &ImageViewer{
		httpClient:  httpClient,
		accessToken: accessToken,
		tempDir:     tempDir,
	}, nil
}

func (iv *ImageViewer) DownloadImage(baseUrl, filename string) (string, error) {
	fmt.Printf("   デバッグ: ディレクトリ確認 - %s\n", iv.tempDir)
	
	// ディレクトリの存在確認
	if _, err := os.Stat(iv.tempDir); os.IsNotExist(err) {
		fmt.Printf("   デバッグ: ディレクトリが存在しないため作成中...\n")
		if err := os.MkdirAll(iv.tempDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create temp directory: %v", err)
		}
		fmt.Printf("   デバッグ: ディレクトリ作成完了 - %s\n", iv.tempDir)
	} else {
		fmt.Printf("   デバッグ: ディレクトリ存在確認済み\n")
	}

	// ファイル拡張子を決定
	ext := filepath.Ext(filename)
	if ext == "" {
		// MIMEタイプから拡張子を推測
		if strings.Contains(baseUrl, "image/") {
			ext = ".jpg" // デフォルト
		}
	}
	fmt.Printf("   デバッグ: 使用する拡張子 - %s\n", ext)

	// 一時ファイルパスを生成
	tempFile := filepath.Join(iv.tempDir, fmt.Sprintf("%d%s", time.Now().UnixNano(), ext))
	fmt.Printf("   デバッグ: 保存先ファイルパス - %s\n", tempFile)

	// 画像をダウンロード
	fmt.Printf("   デバッグ: HTTPリクエスト作成中 - %s\n", baseUrl[:80]+"...")
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 認証ヘッダーを追加
	req.Header.Set("Authorization", "Bearer "+iv.accessToken)
	fmt.Printf("   デバッグ: 認証ヘッダー設定完了\n")

	fmt.Printf("   デバッグ: HTTPリクエスト送信中...\n")
	resp, err := iv.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("   デバッグ: HTTPレスポンス - ステータス: %d, Content-Length: %s\n", 
		resp.StatusCode, resp.Header.Get("Content-Length"))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// ファイルに保存
	fmt.Printf("   デバッグ: ファイル作成中 - %s\n", tempFile)
	file, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	fmt.Printf("   デバッグ: 画像データ書き込み中...\n")
	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %v", err)
	}

	fmt.Printf("   デバッグ: 書き込み完了 - %d バイト\n", bytesWritten)

	// ファイルの存在と権限を確認
	if info, err := os.Stat(tempFile); err == nil {
		fmt.Printf("   デバッグ: ファイル確認済み - サイズ: %d バイト, 権限: %s\n", 
			info.Size(), info.Mode().String())
	} else {
		fmt.Printf("   デバッグ: ファイル確認エラー - %v\n", err)
	}

	return tempFile, nil
}

func (iv *ImageViewer) OpenWithDefaultViewer(imagePath string) error {
	fmt.Printf("   デバッグ: 外部ビューアー起動開始 - OS: %s, ファイル: %s\n", runtime.GOOS, imagePath)
	
	// ファイルの存在確認
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("image file does not exist: %s", imagePath)
	}
	fmt.Printf("   デバッグ: ファイル存在確認済み\n")

	var cmd *exec.Cmd
	var fallbackMsg string

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", imagePath)
		fmt.Printf("   デバッグ: Windows用コマンド実行 - rundll32\n")
	case "darwin":
		cmd = exec.Command("open", imagePath)
		fmt.Printf("   デバッグ: macOS用コマンド実行 - open\n")
	case "linux":
		// Linux環境での複数のビューアーを試行
		viewers := []string{"xdg-open", "eog", "feh", "display", "firefox", "chromium"}
		for _, viewer := range viewers {
			if _, err := exec.LookPath(viewer); err == nil {
				cmd = exec.Command(viewer, imagePath)
				fmt.Printf("   デバッグ: Linux用コマンド実行 - %s\n", viewer)
				break
			}
		}
		if cmd == nil {
			fallbackMsg = fmt.Sprintf("No suitable image viewer found. File saved at: %s", imagePath)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if cmd == nil {
		fmt.Printf("   ℹ️  %s\n", fallbackMsg)
		return nil
	}

	fmt.Printf("   デバッグ: コマンド実行開始...\n")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("   デバッグ: コマンド実行エラー - %v\n", err)
		fmt.Printf("   ℹ️  External viewer failed. File saved at: %s\n", imagePath)
		return nil // エラーとして扱わず、ファイル保存成功として処理
	}
	
	fmt.Printf("   デバッグ: 外部ビューアー起動完了\n")
	return nil
}

func (iv *ImageViewer) DisplayASCII(imagePath string, width int) error {
	fmt.Printf("ASCII Preview of: %s\n", filepath.Base(imagePath))
	
	// 画像ファイルを開く
	file, err := os.Open(imagePath)
	if err != nil {
		return iv.displayPlaceholder(width)
	}
	defer file.Close()
	
	// 画像をデコード
	img, _, err := image.Decode(file)
	if err != nil {
		// HEICなど未対応形式の場合はプレースホルダーを表示
		fmt.Printf("Note: %s format not supported for ASCII preview\n", filepath.Ext(imagePath))
		return iv.displayPlaceholder(width)
	}
	
	// 画像をリサイズ（アスペクト比を維持）
	height := width / 2 // ターミナルでは文字の縦横比を考慮
	resized := resize.Resize(uint(width-4), uint(height), img, resize.Lanczos3)
	
	// ASCII文字のパレット（暗→明）
	palette := " .:-=+*#%@"
	
	fmt.Println("┌" + strings.Repeat("─", width-2) + "┐")
	
	bounds := resized.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		line := "│"
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// ピクセルの輝度を計算
			r, g, b, _ := resized.At(x, y).RGBA()
			// RGBから輝度を計算（0-255）
			gray := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256
			
			// 輝度に基づいてASCII文字を選択
			index := int(gray * float64(len(palette)-1) / 255)
			if index >= len(palette) {
				index = len(palette) - 1
			}
			line += string(palette[index])
		}
		// 行を幅に合わせて調整
		for len(line) < width-1 {
			line += " "
		}
		line += "│"
		fmt.Println(line)
	}
	
	fmt.Println("└" + strings.Repeat("─", width-2) + "┘")
	fmt.Printf("Image: %dx%d pixels\n", bounds.Dx(), bounds.Dy())
	
	return nil
}

func (iv *ImageViewer) displayPlaceholder(width int) error {
	fmt.Println("┌" + strings.Repeat("─", width-2) + "┐")
	
	for i := 0; i < 10; i++ {
		line := "│"
		for j := 0; j < width-4; j++ {
			if (i+j)%3 == 0 {
				line += "█"
			} else if (i+j)%2 == 0 {
				line += "▓"
			} else {
				line += "░"
			}
		}
		line += " │"
		fmt.Println(line)
	}
	
	fmt.Println("└" + strings.Repeat("─", width-2) + "┘")
	fmt.Println("Note: Preview unavailable for this image format. For full image, use --open flag.")
	
	return nil
}

func (iv *ImageViewer) CleanupTempFiles() error {
	// 1時間以上古い一時ファイルを削除
	cutoff := time.Now().Add(-1 * time.Hour)
	
	return filepath.Walk(iv.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // エラーは無視
		}
		
		if !info.IsDir() && info.ModTime().Before(cutoff) {
			os.Remove(path)
		}
		
		return nil
	})
}

func (iv *ImageViewer) GetImageInfo(imagePath string) (map[string]interface{}, error) {
	info, err := os.Stat(imagePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"path":     imagePath,
		"size":     info.Size(),
		"modified": info.ModTime(),
	}, nil
}

// サムネイル用のサイズ調整されたURLを生成
func (iv *ImageViewer) GetThumbnailURL(baseUrl string, width, height int) string {
	if strings.Contains(baseUrl, "googleusercontent.com") {
		// Google Photos の画像リサイズパラメータを追加
		return fmt.Sprintf("%s=w%d-h%d", baseUrl, width, height)
	}
	return baseUrl
}

// 高解像度画像のURLを生成
func (iv *ImageViewer) GetHighResURL(baseUrl string) string {
	if strings.Contains(baseUrl, "googleusercontent.com") {
		// オリジナルサイズまたは高解像度バージョン
		return baseUrl + "=d"
	}
	return baseUrl
}