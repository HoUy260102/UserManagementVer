package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	Id            primitive.ObjectID `bson:"_id"`
	UserId        primitive.ObjectID `bson:"user_id"`
	RefreshToken  string             `bson:"refresh_token"`
	IsRevoked     bool               `bson:"is_revoked"`
	TrustedDevice bool               `bson:"trusted_device"`
	DeviceId      string             `bson:"device_id"`
	CreatedAt     time.Time          `bson:"created_at"`
	ExpiresAt     time.Time          `bson:"expires_at"`
	ApprovedToken string             `bson:"approved_token"`
}
