package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name,omitempty"`
	Email     string             `bson:"email,omitempty"`
	Password  string             `bson:"password,omitempty"`
	Phone     string             `bson:"phone,omitempty"`
	Dob       time.Time          `bson:"dob,omitempty"`
	ImageUrl  string             `bson:"image_url,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	CreatedBy string             `bson:"created_by,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	UpdatedBy string             `bson:"updated_by,omitempty"`
	DeletedAt time.Time          `bson:"deleted_at,omitempty"`
	DeletedBy string             `bson:"deleted_by,omitempty"`
}
