package collections

import (
	"UserManagementVer/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionCollection struct {
	collection *mongo.Collection
}

func NewSessionCollection(collection *mongo.Collection) *SessionCollection {
	return &SessionCollection{collection}
}

func (sessionCollection *SessionCollection) FindOne(ctx context.Context, filter bson.M) (models.Session, error) {
	var session models.Session
	err := sessionCollection.collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func (sessionCollection *SessionCollection) Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]models.Session, error) {
	var sessions []models.Session

	cursor, err := sessionCollection.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (sessionCollection *SessionCollection) FindAndUpdate(ctx context.Context, session models.Session) (models.Session, error) {
	filter := bson.M{
		"user_id":   session.UserId,
		"device_id": session.DeviceId,
	}

	update := bson.M{
		"$set": bson.M{
			"refresh_token":  session.RefreshToken,
			"created_at":     session.CreatedAt,
			"expires_at":     session.ExpiresAt,
			"is_revoked":     session.IsRevoked,
			"trusted_device": session.TrustedDevice,
			"approved_token": session.ApprovedToken,
		},
		"$setOnInsert": bson.M{ // chỉ áp dụng khi insert mới
			"user_id":   session.UserId,
			"device_id": session.DeviceId,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true). // cho phép insert nếu không có
		SetReturnDocument(options.After)

	var updated models.Session
	res := sessionCollection.collection.FindOneAndUpdate(ctx, filter, update, opts)
	if res.Err() != nil {
		return updated, res.Err()
	}
	err := res.Decode(&updated)
	return updated, err
}

func (sessionCollection *SessionCollection) DeleteSession(ctx context.Context, filter bson.M) error {
	_, err := sessionCollection.collection.DeleteOne(ctx, filter)
	return err
}
