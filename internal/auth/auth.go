package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	credentialsFile    = "credentials.json"
	tokenFile          = "token.json"
	credentialsDirPerm = 0700
	tokenFilePerm      = 0600
)

var scopes = []string{
	docs.DocumentsScope,
	drive.DriveScope,
}

// GetCredentialsPath returns the path to credentials directory (same as gdrive)
func GetCredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gdrive")
}

// GetClient retrieves an OAuth2 HTTP client
func GetClient(ctx context.Context) (*http.Client, error) {
	credPath := filepath.Join(GetCredentialsPath(), credentialsFile)
	tokenPath := filepath.Join(GetCredentialsPath(), tokenFile)

	b, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file %s: %w\n"+
			"See README.md for setup instructions", credPath, err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	token, err := tokenFromFile(tokenPath)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokenPath, token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: unable to save token: %v\n", err)
		}
	}

	return config.Client(ctx, token), nil
}

// GetDocsService creates an authenticated Docs service
func GetDocsService(ctx context.Context) (*docs.Service, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Docs service: %w", err)
	}

	return service, nil
}

// GetDriveService creates an authenticated Drive service
func GetDriveService(ctx context.Context) (*drive.Service, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %w", err)
	}

	return service, nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Printf("Enter authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Fprintf(os.Stderr, "Saving credentials to: %s\n", path)

	if err := os.MkdirAll(filepath.Dir(path), credentialsDirPerm); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, tokenFilePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}
