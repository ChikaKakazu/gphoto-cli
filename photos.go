package main

import (
	"context"
	"fmt"
	"net/http"

	gphotos "github.com/gphotosuploader/google-photos-api-client-go/v3"
)

type PhotosClient struct {
	client *gphotos.Client
}

func NewPhotosClient(httpClient *http.Client) (*PhotosClient, error) {
	client, err := gphotos.NewClient(httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Photos client: %v", err)
	}

	return &PhotosClient{client: client}, nil
}

func (pc *PhotosClient) ListAlbums(ctx context.Context) error {
	fmt.Println("Fetching albums from Google Photos...")

	albums, err := pc.client.Albums.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list albums: %v", err)
	}

	if len(albums) == 0 {
		fmt.Println("No albums found in your Google Photos.")
		return nil
	}

	fmt.Printf("Found %d album(s):\n\n", len(albums))

	for i, album := range albums {
		fmt.Printf("%d. %s\n", i+1, album.Title)
		fmt.Printf("   ID: %s\n", album.ID)
		if album.ProductURL != "" {
			fmt.Printf("   URL: %s\n", album.ProductURL)
		}
		fmt.Println()
	}

	return nil
}

func (pc *PhotosClient) ListMediaItems(ctx context.Context, limit int) error {
	fmt.Printf("This feature is currently under development.\n")
	fmt.Printf("Would list up to %d media items from Google Photos.\n", limit)
	return nil
}
