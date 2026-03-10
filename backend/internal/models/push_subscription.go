package models

import (
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Endpoint  string    `db:"endpoint" json:"endpoint"`
	Auth      string    `db:"auth" json:"auth"`
	P256dh    string    `db:"p256dh" json:"p256dh"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type PushSubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Auth     string `json:"auth" binding:"required"`
	P256dh   string `json:"p256dh" binding:"required"`
}
