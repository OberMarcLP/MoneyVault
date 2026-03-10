package services

import (
	"encoding/json"
	"log"

	"moneyvault/internal/config"
	"moneyvault/internal/models"
	"moneyvault/internal/repositories"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
)

type PushService struct {
	repo       *repositories.PushRepository
	vapidPub   string
	vapidPriv  string
}

func NewPushService(repo *repositories.PushRepository, cfg *config.Config) *PushService {
	return &PushService{
		repo:      repo,
		vapidPub:  cfg.VAPIDPublicKey,
		vapidPriv: cfg.VAPIDPrivateKey,
	}
}

func (s *PushService) Subscribe(userID uuid.UUID, req models.PushSubscribeRequest) error {
	sub := &models.PushSubscription{
		ID:       uuid.New(),
		UserID:   userID,
		Endpoint: req.Endpoint,
		Auth:     req.Auth,
		P256dh:   req.P256dh,
	}
	return s.repo.Subscribe(sub)
}

func (s *PushService) Unsubscribe(userID uuid.UUID, endpoint string) error {
	return s.repo.Unsubscribe(userID, endpoint)
}

func (s *PushService) GetVAPIDPublicKey() string {
	return s.vapidPub
}

// SendToUser sends a push notification to all of a user's subscriptions.
func (s *PushService) SendToUser(userID uuid.UUID, title, body, url string) {
	if s.vapidPub == "" || s.vapidPriv == "" {
		return // Push not configured
	}

	subs, err := s.repo.GetByUser(userID)
	if err != nil || len(subs) == 0 {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"body":  body,
		"url":   url,
	})

	for _, sub := range subs {
		wp := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.Auth,
				P256dh: sub.P256dh,
			},
		}

		resp, err := webpush.SendNotification(payload, wp, &webpush.Options{
			VAPIDPublicKey:  s.vapidPub,
			VAPIDPrivateKey: s.vapidPriv,
			Subscriber:      "mailto:admin@moneyvault.app",
		})
		if err != nil {
			log.Printf("Push failed for %s: %v", sub.Endpoint[:40], err)
			// Remove invalid subscription (410 Gone or 404)
			if resp != nil && (resp.StatusCode == 410 || resp.StatusCode == 404) {
				_ = s.repo.Delete(sub.ID)
			}
			continue
		}
		resp.Body.Close()
	}
}
