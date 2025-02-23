package readunread

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

type Oauth2Srv struct {
	srv *http.Server
	ch  chan string
	cfg *oauth2.Config
	ctx context.Context
}

// Create a new HTTP server on the provided addr. This immediately starts
// the server in a goroutine.
func NewServer(ctx context.Context, config *oauth2.Config, addr string) Oauth2Srv {
	ch := make(chan string)
	srv := &http.Server{Addr: addr, Handler: oauthHandler(ch)}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
		log.Println("Server stopped")
	}()

	return Oauth2Srv{
		srv: srv,
		ch:  ch,
		cfg: config,
		ctx: ctx,
	}
}

func (s *Oauth2Srv) WaitForJwt() (*oauth2.Token, error) {
	authCode := <-s.ch
	return s.cfg.Exchange(s.ctx, authCode)
}

func (s *Oauth2Srv) Stop() error {
	close(s.ch)
	if err := s.srv.Shutdown(s.ctx); err != nil {
		return err
	}
	return nil
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
