package unreadread

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	oauth2srv "github.com/KevinMidkiff/unreadread/internal/oauth2srv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func GetGmailService(credsFile string) (*gmail.Service, error) {
	log.Printf("Loading configuration from %s\n", credsFile)
	cfg, err := loadOauth2Config(credsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load oauth2 config: %v", err)
	}

	srv, err := initGmailService(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gmail service: %v", err)
	}

	return srv, nil
}

func loadOauth2Config(credsFile string) (*oauth2.Config, error) {
	contents, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(contents, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to init google oauth2 config: %v", err)
	}

	config.RedirectURL = "http://localhost:8081"

	return config, nil
}

func initGmailService(config *oauth2.Config) (*gmail.Service, error) {
	client := getHttpClient(config)
	return gmail.New(client)
}

// Retrieve a token, saves the token, then returns the generated client.
func getHttpClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/gmail.modify"}
	authURL := config.AuthCodeURL(
		"state-token", oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("scope", strings.Join(scopes, " ")))
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	srv := oauth2srv.NewServer(context.TODO(), config, ":8081")
	tok, err := srv.WaitForJwt()
	if err != nil {
		log.Fatalf("Failed to get JWT: %v", err)
	}

	return tok
}

// Retrieves a token from a local file.
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

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
