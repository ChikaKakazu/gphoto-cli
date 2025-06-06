package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PickerSession struct {
	Name            string `json:"name"`
	PickerUri       string `json:"pickerUri"`
	MediaItemsSet   bool   `json:"mediaItemsSet"`
	ID              string `json:"id"`
}

type MediaFile struct {
	BaseUrl             string                 `json:"baseUrl"`
	MimeType            string                 `json:"mimeType"`
	Filename            string                 `json:"filename"`
	MediaFileMetadata   MediaFileMetadata      `json:"mediaFileMetadata"`
}

type MediaFileMetadata struct {
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	CameraMake  string        `json:"cameraMake"`
	CameraModel string        `json:"cameraModel"`
	PhotoMetadata PhotoMetadata `json:"photoMetadata"`
}

type PhotoMetadata struct {
	FocalLength     float64 `json:"focalLength"`
	ApertureFNumber float64 `json:"apertureFNumber"`
	IsoEquivalent   int     `json:"isoEquivalent"`
	ExposureTime    string  `json:"exposureTime"`
}

type MediaItem struct {
	ID          string    `json:"id"`
	CreateTime  string    `json:"createTime"`
	Type        string    `json:"type"`
	MediaFile   MediaFile `json:"mediaFile"`
}

type MediaItemsResponse struct {
	MediaItems []MediaItem `json:"mediaItems"`
}

type PickerClient struct {
	httpClient  *http.Client
	accessToken string
}

func NewPickerClient(httpClient *http.Client, accessToken string) *PickerClient {
	return &PickerClient{
		httpClient:  httpClient,
		accessToken: accessToken,
	}
}

func (pc *PickerClient) CreateSession(ctx context.Context) (*PickerSession, error) {
	url := "https://photospicker.googleapis.com/v1/sessions"
	
	// セッション作成リクエスト（空のオブジェクト）
	reqBody := map[string]interface{}{}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+pc.accessToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	// レスポンス全体を読み取ってデバッグ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	fmt.Printf("DEBUG: Session creation response: %s\n", string(body))
	
	var session PickerSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	// セッション名が空の場合はIDを使用
	if session.Name == "" && session.ID != "" {
		session.Name = fmt.Sprintf("sessions/%s", session.ID)
	}
	
	fmt.Printf("DEBUG: Session name: %s\n", session.Name)
	
	return &session, nil
}

func (pc *PickerClient) GetSession(ctx context.Context, sessionName string) (*PickerSession, error) {
	url := fmt.Sprintf("https://photospicker.googleapis.com/v1/%s", sessionName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+pc.accessToken)
	
	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	var session PickerSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return &session, nil
}

func (pc *PickerClient) ListMediaItems(ctx context.Context, sessionName string) ([]MediaItem, error) {
	// sessionNameから sessionId を抽出 (sessions/xxxxx-xxxx -> xxxxx-xxxx)
	sessionId := sessionName
	if strings.HasPrefix(sessionName, "sessions/") {
		sessionId = strings.TrimPrefix(sessionName, "sessions/")
	}
	
	// 正しいエンドポイント: /v1/mediaItems?sessionId=xxx
	url := fmt.Sprintf("https://photospicker.googleapis.com/v1/mediaItems?sessionId=%s", sessionId)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+pc.accessToken)
	
	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	
	// レスポンス全体を読み取ってデバッグ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	fmt.Printf("DEBUG: MediaItems response: %s\n", string(body))
	
	var response MediaItemsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return response.MediaItems, nil
}

func (pc *PickerClient) WaitForSelection(ctx context.Context, sessionName string) error {
	fmt.Println("ユーザーの写真選択を待っています...")
	
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(10 * time.Minute)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("写真選択がタイムアウトしました")
		case <-ticker.C:
			session, err := pc.GetSession(ctx, sessionName)
			if err != nil {
				return fmt.Errorf("セッション取得エラー: %v", err)
			}
			
			if session.MediaItemsSet {
				fmt.Println("写真が選択されました！")
				return nil
			}
		}
	}
}