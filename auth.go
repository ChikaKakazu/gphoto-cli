package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const (
	tokenFile = "token.json"
)

func getGoogleConfig() (*oauth2.Config, error) {
	// 設定ファイルから読み込み（優先）
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// 設定が完了していない場合はエラー
	if config.GoogleClientID == "" || config.GoogleClientSecret == "" {
		return nil, fmt.Errorf("Google OAuth credentials not configured. Please run: ./gphoto-cli setup")
	}

	// 環境変数からの上書き（オプション）
	if err := godotenv.Load(); err == nil {
		if envClientID := os.Getenv("GOOGLE_CLIENT_ID"); envClientID != "" {
			config.GoogleClientID = envClientID
		}
		if envClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET"); envClientSecret != "" {
			config.GoogleClientSecret = envClientSecret
		}
		if envRedirectURI := os.Getenv("GOOGLE_REDIRECT_URI"); envRedirectURI != "" {
			config.GoogleRedirectURI = envRedirectURI
		}
		if envScope := os.Getenv("GOOGLE_SCOPE"); envScope != "" {
			config.GoogleScope = envScope
		}
		if envAuthMethod := os.Getenv("AUTH_METHOD"); envAuthMethod != "" {
			config.AuthMethod = envAuthMethod
		}
	}

	// OAuth2設定を構築
	oauthConfig := &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  config.GoogleRedirectURI,
		Scopes:       []string{config.GoogleScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	return oauthConfig, nil
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// 設定ファイルから認証方式を取得
	appConfig, err := loadConfig()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		fmt.Println("自動認証方式を使用します")
		return getTokenWithLocalServer(config)
	}

	// 環境変数での上書きもチェック
	authMethod := appConfig.AuthMethod
	if envAuthMethod := os.Getenv("AUTH_METHOD"); envAuthMethod != "" {
		authMethod = envAuthMethod
	}

	switch authMethod {
	case "server":
		fmt.Println("自動認証方式を使用します (ローカルサーバー)")
		return getTokenWithLocalServer(config)
	case "oob":
		fmt.Println("手動認証方式を使用します (認証コード入力)")
		return getTokenManually(config)
	default:
		fmt.Printf("不明な認証方式: %s\n", authMethod)
		fmt.Println("自動認証方式を使用します")
		return getTokenWithLocalServer(config)
	}
}

func getTokenWithLocalServer(config *oauth2.Config) *oauth2.Token {
	codeCh := make(chan string)
	state := "state-token"
	
	// 設定からポート番号を決定
	port := ":8080" // デフォルト
	if config.RedirectURL != "" && config.RedirectURL != "urn:ietf:wg:oauth:2.0:oob" {
		// 既にconfigに設定されているRedirectURLを使用
	}
	
	// ローカルサーバーを起動
	server := &http.Server{Addr: port}
	
	http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in request", http.StatusBadRequest)
			return
		}
		
		fmt.Fprintf(w, "<html><body><h1>認証が完了しました！</h1><p>このタブを閉じて、ターミナルに戻ってください。</p></body></html>")
		
		// コードをチャネルに送信
		go func() {
			codeCh <- code
		}()
	})
	
	// サーバーを別ゴルーチンで起動
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ローカルサーバーエラー: %v", err)
			log.Println("手動認証方式に切り替えてください")
		}
	}()
	
	// 認証URLを生成
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	fmt.Printf("ブラウザで以下のURLを開いて認証を行ってください:\n%v\n\n", authURL)
	fmt.Println("認証完了まで待機中...")
	
	// 認証コードを待機（タイムアウト付き）
	var code string
	select {
	case code = <-codeCh:
		fmt.Println("認証コードを受信しました")
	case <-time.After(3 * time.Minute):
		fmt.Println("ローカルサーバー認証がタイムアウトしました")
		// サーバーを停止
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		// 手動認証にフォールバック
		return getTokenManually(config)
	}
	
	// サーバーを停止
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	
	// トークンを取得
	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("トークンの取得に失敗しました: %v", err)
	}
	
	return tok
}

func getTokenManually(config *oauth2.Config) *oauth2.Token {
	// デスクトップアプリケーション用のOOB (Out of Band) フロー
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	
	fmt.Printf("\n=== 手動認証方式 ===\n")
	fmt.Printf("1. ブラウザで以下のURLを開いてください:\n%v\n\n", authURL)
	fmt.Println("2. Google認証を完了してください")
	fmt.Println("3. 表示された認証コードをコピーしてください")
	
	var authCode string
	fmt.Print("\n認証コードを入力してください: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("認証コードの読み取りに失敗しました: %v", err)
	}
	
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("トークンの取得に失敗しました: %v", err)
	}
	
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getClient(config *oauth2.Config) *http.Client {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getAccessToken(config *oauth2.Config) (string, error) {
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return tok.AccessToken, nil
}