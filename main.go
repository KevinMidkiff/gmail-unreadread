package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	stopwatch "github.com/KevinMidkiff/readunread/internal/stopwatch"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const CREDENTIALS_JSON = "credentials.json"

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
			log.Fatalf("%v", err)
		}

		if r.NextPageToken == "" {
			break
		}

		modifySw.Start()
		if err = markRead(srv, r.Messages); err != nil {
			log.Fatalf("%v", err)
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
	r, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get next batch of emails: %v", err)
	}
	return r, nil
}

func markRead(srv *gmail.Service, msgs []*gmail.Message) error {
	mod := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	for _, msg := range msgs {
		_, err := srv.Users.Messages.Modify("me", msg.Id, mod).Do()
		if err != nil {
			return fmt.Errorf("error modifying message to be read: %v", err)
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
	config, err := google.ConfigFromJSON(contents, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to init google oauth2 config: %v", err)
	}

	config.RedirectURL = "http://localhost:8081"

	return config, nil
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

func oauthHandler(ch chan<- string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			log.Printf("Error: failed to parse URL: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			fmt.Fprintf(w, "ERROR: GET request missing 'code' in URL")
			return
		}

		query := u.Query()
		code := query.Get("code")
		if code == "" {
			log.Println("Error: Empty code value in URL")
			http.Error(w, "Bad Request", http.StatusBadRequest)
			fmt.Fprintf(w, "ERROR: GET request 'code' parameter is empty")
			return
		}

		fmt.Fprintf(w, "success, you can close this window")
		ch <- code
	}
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	scopes := []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/gmail.modify"}
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("scope", strings.Join(scopes, " ")))
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	messages := make(chan string)
	defer close(messages)

	server := &http.Server{Addr: ":8081", Handler: oauthHandler(messages)}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
		log.Println("Server stopped")
	}()

	authCode := <-messages
	if err := server.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Failed to stop oauth2 server: %v", err)
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
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
