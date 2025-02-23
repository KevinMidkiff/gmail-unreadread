package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	stopwatch "github.com/KevinMidkiff/readunread/internal/stopwatch"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	CREDENTIALS_JSON = "credentials.json"
	PRINT_BUF        = 1000
)

func main() {
	config, err := loadOauth2Config(CREDENTIALS_JSON)
	if err != nil {
		log.Fatalf("Failed to load oauth2 config: %v", err)
	}

	srv, err := getGmailService(config)
	if err != nil {
		log.Fatalf("Failed to get gmail service instance: %v", err)
	}

	pageToken := ""
	numUnmarked := 0
	listSw := stopwatch.New()
	modifySw := stopwatch.New()
	totalSw := stopwatch.New()

	log.Println("Starting to query unread messages")
	totalSw.Start()
	for {
		listSw.Start()
		r, err := nextMessages(srv, pageToken)
		if err != nil {
			log.Fatalf("Failed to retrieve next set of messages: %v", err)
		}

		if r.NextPageToken == "" {
			break
		}

		modifySw.Start()
		if err = markRead(srv, r.Messages); err != nil {
			log.Fatalf("Failed to mark messages as read: %v", err)
		}

		pageToken = r.NextPageToken
		numMsgs := len(r.Messages)
		numUnmarked += numMsgs

		log.Printf("Finished batch of %d msgs - total: %d, list-elapsed: %s, modify-elapsed: %s, total-elapsed: %s\n",
			numMsgs, numUnmarked, listSw.Elapsed(), modifySw.Elapsed(), totalSw.Elapsed())
	}

	log.Printf("Total number of messages marked read: %d, total_elapsed: %s", numUnmarked, totalSw.Elapsed())
}

func nextMessages(srv *gmail.Service, pageToken string) (*gmail.ListMessagesResponse, error) {
	req := srv.Users.Messages.List("me").Q("is:unread")
	if pageToken != "" {
		req.PageToken(pageToken)
	}
	return req.Do()
}

func markRead(srv *gmail.Service, msgs []*gmail.Message) error {
	mod := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	for _, msg := range msgs {
		_, err := srv.Users.Messages.Modify("me", msg.Id, mod).Do()
		if err != nil {
			return fmt.Errorf("Error modifying message to be read: %v", err)
		}
	}
	return nil
}

func loadOauth2Config(credsFile string) (*oauth2.Config, error) {
	log.Printf("Loading configuration from %s\n", credsFile)
	contents, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json.
	return google.ConfigFromJSON(contents, gmail.GmailReadonlyScope)
}

func getGmailService(config *oauth2.Config) (*gmail.Service, error) {
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
	scopes := []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/gmail.modify"}
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("scope", strings.Join(scopes, " ")))
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
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
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
