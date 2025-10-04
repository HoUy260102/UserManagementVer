package collections

import (
	"UserManagementVer/models"
	"UserManagementVer/utils"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountCollection struct {
	collection *mongo.Collection
}

func NewAccountCollection(collection *mongo.Collection) *AccountCollection {
	return &AccountCollection{
		collection: collection,
	}
}

func (a *AccountCollection) Create(ctx context.Context, account models.Account) error {
	var (
		err error
	)
	account.Password, err = utils.HashPassword(account.Password)
	account.CreatedAt = time.Now()
	if err != nil {
		return err
	}
	_, err = a.collection.InsertOne(ctx, account)
	return err
}

func (a *AccountCollection) GetAccountById(ctx context.Context, objectId primitive.ObjectID) (models.Account, error) {

	var (
		account models.Account
		err     error
	)
	fmt.Println(objectId.Hex())
	err = a.collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&account)
	if err != nil {
		return account, err
	}
	return account, nil
}

func (a *AccountCollection) Find(ctx context.Context, filter bson.M) (models.Account, error) {
	var account models.Account
	err := a.collection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		return models.Account{}, err
	}
	return account, nil
}

func (a *AccountCollection) FindAll(ctx context.Context, filter bson.M) ([]models.Account, error) {
	var accounts []models.Account
	cursor, err := a.collection.Find(ctx, filter)
	if err != nil {
		return accounts, err
	}
	err = cursor.All(ctx, &accounts)
	if err != nil {
		return accounts, err
	}
	return accounts, nil
}

func (a *AccountCollection) Update(ctx context.Context, filter bson.M, update bson.M) error {
	res, err := a.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("Không có tài liệu được update")
	}
	return nil
}

func (a *AccountCollection) DeleteIndex(indexName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, _ := a.collection.Indexes().List(ctx)

	for cursor.Next(ctx) {
		var indexInfo map[string]interface{}
		if err := cursor.Decode(&indexInfo); err != nil {
			panic(err)
		}

		// So sánh tên index
		if name, ok := indexInfo["name"].(string); ok && name == indexName {
			fmt.Println("Delete successful")
			_, err := a.collection.Indexes().DropOne(ctx, indexName) // tên mặc định
			if err != nil {
				return
			}
			return
		}
	}
}

func (a *AccountCollection) UpdateIndex(ttl int, indexName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{indexName: 1},
		Options: options.Index().SetExpireAfterSeconds(int32(ttl*24*60) * 60), // TTL mới
	}

	_, err := a.collection.Indexes().CreateOne(ctx, indexModel)
	return err

}

func (a *AccountCollection) SearchByText(keyword string) ([]models.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result []models.Account
	keyword = strings.ToLower(keyword)

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"$text", bson.D{
				{"$search", keyword},
			}},
		}}},
	}

	cusor, err := a.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return result, err
	}

	defer cusor.Close(ctx)
	for cusor.Next(ctx) {
		var account models.Account
		if err := cusor.Decode(&account); err != nil {
			return result, err
		}
		result = append(result, account)
	}
	
	return result, nil
}
