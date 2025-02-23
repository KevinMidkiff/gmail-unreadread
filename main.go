package main

import (
	"log"

	stopwatch "github.com/KevinMidkiff/unreadread/internal/stopwatch"
	unreadread "github.com/KevinMidkiff/unreadread/internal/unreadread"
)

const (
	CREDENTIALS_JSON = "credentials.json"
	NUM_WORKERS      = 12
	NUM_QUEUED_MSGS  = 1024
)

func main() {
	log.Printf("Initializing with %d workers, %d queue size, using creds in '%s'", NUM_WORKERS, NUM_QUEUED_MSGS, CREDENTIALS_JSON)
	r, err := unreadread.NewUnreadRead(NUM_WORKERS, NUM_QUEUED_MSGS, CREDENTIALS_JSON)
	if err != nil {
		log.Fatalf("%v", err)
	}

	sw := stopwatch.New()
	sw.Start()
	log.Println("Processing messages")
	r.ProcessUnreadMsgs()
	log.Printf("Finished - total elapsed: %s", sw.Elapsed())
}
