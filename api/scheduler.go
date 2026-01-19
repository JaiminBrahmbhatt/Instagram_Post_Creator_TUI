package api

import (
	"fmt"
	"log"
	"time"

	"github.com/JaiminBrahmbhatt/Instagram_Post_Creator_TUI/db"
)

type Scheduler struct {
	Client      *Client
	DB          *db.Database
	ReportChan  chan string
	TriggerChan chan struct{}
}

func NewScheduler(database *db.Database, client *Client) *Scheduler {
	return &Scheduler{
		Client:      client,
		DB:          database,
		ReportChan:  make(chan string, 100),
		TriggerChan: make(chan struct{}, 1),
	}
}

func (s *Scheduler) CheckAndPublish() {
	posts, err := s.DB.GetScheduledPosts()
	if err != nil {
		log.Printf("Scheduler error: %v", err)
		return
	}

	for _, p := range posts {
		// Mark as publishing immediately to prevent double-processing
		if err := s.DB.MarkPostStatus(p.ID, db.StatusPublishing); err != nil {
			log.Printf("Failed to mark post %d as publishing: %v", p.ID, err)
			continue
		}
		go s.PublishPost(p.ID, p.Caption)
	}
}

func (s *Scheduler) report(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	log.Println(msg) // Now safely writing to debug.log
	s.ReportChan <- msg
}

// markFailed is now handled via DB.MarkPostStatus(id, db.StatusFailed) directly in publisher.go

func (s *Scheduler) Trigger() {
	select {
	case s.TriggerChan <- struct{}{}:
	default:
		// Already triggered
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CheckAndPublish()
			case <-s.TriggerChan:
				s.CheckAndPublish()
			}
		}
	}()
}
