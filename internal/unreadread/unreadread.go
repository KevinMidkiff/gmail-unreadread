package unreadread

import (
	"fmt"
	"log"
	"sync"

	gmailsrv "github.com/KevinMidkiff/unreadread/internal/gmailsrv"
	stopwatch "github.com/KevinMidkiff/unreadread/internal/stopwatch"
	"google.golang.org/api/gmail/v1"
)

type UnreadRead struct {
	srv    *gmail.Service
	msgIds chan string
	wg     *sync.WaitGroup
}

func NewUnreadRead(numWorkers int, numQueuedMsgs int, credsFile string) (*UnreadRead, error) {
	srv, err := gmailsrv.GetGmailService(credsFile)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	msgIds := make(chan string, numQueuedMsgs)

	for i := range numWorkers {
		wg.Add(1)
		go worker(i, srv, msgIds, &wg)
	}

	return &UnreadRead{srv, msgIds, &wg}, nil
}

func (u *UnreadRead) ProcessUnreadMsgs() error {
	pageToken := ""
	numUnmarked := 0
	sw := stopwatch.New()
	totalElapsed := stopwatch.New()
	totalElapsed.Start()

	for {
		sw.Start()
		r, err := u.nextMessages(pageToken)
		if err != nil {
			return fmt.Errorf("failed to list next messages: %v", err)
		}
		listElapsed := sw.Elapsed()
		sw.Start()
		for _, msg := range r.Messages {
			u.msgIds <- msg.Id
		}
		enqueueElapsed := sw.Elapsed()

		if r.NextPageToken == "" {
			break
		}

		pageToken = r.NextPageToken
		numUnmarked += len(r.Messages)

		log.Printf("Enqueued new batch (msgs: %d, total-elapsed: %s, list-elapsed: %s, enqueue-elapsed: %s)",
			numUnmarked, totalElapsed.Elapsed(), listElapsed, enqueueElapsed)
	}

	return nil
}

func worker(id int, srv *gmail.Service, msgIds <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	numProcessed := 0
	sw := stopwatch.New()
	mod := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}
	sw.Start()

	log.Printf("[job:%d] Processing messages", id)
	for msg := range msgIds {
		_, err := srv.Users.Messages.Modify("me", msg, mod).Do()
		if err != nil {
			log.Printf("Error: modifying message to be read: %v", err)
		}
		numProcessed++
		if numProcessed%10 == 0 || numProcessed == 1 {
			elapsed := sw.Elapsed()
			avg := float64(numProcessed) / elapsed.Seconds()
			log.Printf("[job:%d] Marked %d messages read (avg: %f, elapsed: %s)", id, numProcessed, avg, elapsed)
		}
	}
	log.Printf("[job:%d] Done processing messages", id)
}

func (u *UnreadRead) nextMessages(pageToken string) (*gmail.ListMessagesResponse, error) {
	req := u.srv.Users.Messages.List("me").Q("is:unread")
	if pageToken != "" {
		req.PageToken(pageToken)
	}
	r, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get next batch of emails: %v", err)
	}
	return r, nil
}
