package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GoogleClientID     string `yaml:"google_client_id"`
	GoogleClientSecret string `yaml:"google_client_secret"`
	GoogleRedirectURI  string `yaml:"google_redirect_uri"`
	GoogleScope        string `yaml:"google_scope"`
	AuthMethod         string `yaml:"auth_method"`
}

// デフォルト設定
func getDefaultConfig() *Config {
	return &Config{
		GoogleRedirectURI: "http://localhost:8080/auth/callback",
		GoogleScope:       "https://www.googleapis.com/auth/photospicker.mediaitems.readonly",
		AuthMethod:        "server",
	}
}

// 設定ファイルのパスを取得
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	
	configDir := filepath.Join(homeDir, ".gphoto-cli")
	
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}
	
	return filepath.Join(configDir, "config.yaml"), nil
}

// 設定を読み込み
func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	
	// ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	// デフォルト値を設定
	if config.GoogleRedirectURI == "" {
		config.GoogleRedirectURI = "http://localhost:8080/auth/callback"
	}
	if config.GoogleScope == "" {
		config.GoogleScope = "https://www.googleapis.com/auth/photospicker.mediaitems.readonly"
	}
	if config.AuthMethod == "" {
		config.AuthMethod = "server"
	}
	
	return config, nil
}

// 設定を保存
func saveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	
	return nil
}

// 対話式セットアップ
func runInteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("🔧 gphoto-cli セットアップ")
	fmt.Println("=====================================")
	fmt.Println()
	
	// Google Cloud Console のセットアップ手順を案内
	fmt.Println("📋 Google Cloud Console でのセットアップが必要です:")
	fmt.Println()
	fmt.Println("1. Google Cloud Console (https://console.cloud.google.com/) にアクセス")
	fmt.Println("2. 新しいプロジェクトを作成または既存のプロジェクトを選択")
	fmt.Println("3. APIs & Services > Credentials で 'OAuth 2.0 Client ID' を作成")
	fmt.Println("   - アプリケーションの種類: デスクトップアプリケーション")
	fmt.Println("   - 承認済みのリダイレクト URI: http://localhost:8080/auth/callback")
	fmt.Println("4. クライアント ID とクライアント シークレットをメモ")
	fmt.Println()
	fmt.Println("💡 注意: Google Photos Picker APIは特別な有効化は不要で、")
	fmt.Println("   OAuth認証のみで利用できます。")
	fmt.Println()
	
	fmt.Print("準備ができたら Enter キーを押してください...")
	reader.ReadLine()
	fmt.Println()
	
	config := getDefaultConfig()
	
	// クライアントIDの入力
	fmt.Print("Google Client ID を入力してください: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read client ID: %v", err)
	}
	config.GoogleClientID = strings.TrimSpace(clientID)
	
	if config.GoogleClientID == "" {
		return fmt.Errorf("Client ID は必須です")
	}
	
	// クライアントシークレットの入力
	fmt.Print("Google Client Secret を入力してください: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read client secret: %v", err)
	}
	config.GoogleClientSecret = strings.TrimSpace(clientSecret)
	
	if config.GoogleClientSecret == "" {
		return fmt.Errorf("Client Secret は必須です")
	}
	
	// 認証方式の選択
	fmt.Println()
	fmt.Println("認証方式を選択してください:")
	fmt.Println("1. 自動認証 (推奨): ローカルサーバーを使用")
	fmt.Println("2. 手動認証: 認証コードを手動で入力")
	fmt.Print("選択 (1 または 2) [1]: ")
	
	authChoice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read auth method: %v", err)
	}
	authChoice = strings.TrimSpace(authChoice)
	
	if authChoice == "" || authChoice == "1" {
		config.AuthMethod = "server"
	} else if authChoice == "2" {
		config.AuthMethod = "oob"
		config.GoogleRedirectURI = "urn:ietf:wg:oauth:2.0:oob"
	} else {
		fmt.Println("無効な選択です。自動認証を使用します。")
		config.AuthMethod = "server"
	}
	
	// 設定を保存
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}
	
	configPath, _ := getConfigPath()
	
	fmt.Println()
	fmt.Println("✅ セットアップが完了しました!")
	fmt.Printf("設定ファイル: %s\n", configPath)
	fmt.Println()
	fmt.Println("🚀 次のコマンドで Google Photos にアクセスできます:")
	fmt.Println("   ./gphoto-cli picker")
	fmt.Println()
	
	return nil
}

// 設定の確認
func runConfigShow() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}
	
	configPath, _ := getConfigPath()
	
	fmt.Printf("📍 設定ファイル: %s\n", configPath)
	fmt.Println()
	fmt.Printf("Google Client ID: %s\n", maskString(config.GoogleClientID))
	fmt.Printf("Google Client Secret: %s\n", maskString(config.GoogleClientSecret))
	fmt.Printf("Redirect URI: %s\n", config.GoogleRedirectURI)
	fmt.Printf("認証方式: %s\n", config.AuthMethod)
	fmt.Printf("OAuth Scope: %s\n", config.GoogleScope)
	
	return nil
}

// 設定のリセット
func runConfigReset() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	
	// 設定ファイルを削除
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %v", err)
	}
	
	// トークンファイルも削除
	if err := os.Remove(tokenFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove token file: %v\n", err)
	}
	
	fmt.Println("✅ 設定がリセットされました")
	fmt.Println("再度セットアップを行うには: ./gphoto-cli setup")
	
	return nil
}

// 文字列をマスク表示
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// 設定が完了しているかチェック
func isConfigured() bool {
	config, err := loadConfig()
	if err != nil {
		return false
	}
	
	return config.GoogleClientID != "" && config.GoogleClientSecret != ""
}