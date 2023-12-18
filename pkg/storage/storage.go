package storage

import (
	"challenge/config"
	"challenge/internal/tweet"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Storage struct {
	client *mongo.Client
	db     *mongo.Database
}

func New(ctx context.Context, config config.Configuration) *Storage {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	if config.GetString("url-db", "") == "" {
		log.Fatal("config url-db is invalid")
	}
	opts := options.Client().ApplyURI("mongodb+srv://juliangonzalofernandez:wRscvt6xKM5ABVC5@clusterchallenge.cfzqmlt.mongodb.net/?retryWrites=true&w=majority").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Database("tweet").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		log.Fatalf("do a ping in db: %w", err)
		panic(err)
	}
	color.Green("You successfully connected to MongoDB!")

	return &Storage{
		client: client,
		db:     client.Database("tweet"),
	}
}

func (s *Storage) CreateUser(ctx context.Context, userID string) (*tweet.User, error) {
	user := &tweet.User{UserID: userID}

	_, err := s.db.Collection("users").InsertOne(ctx, bson.D{
		{Key: "user_id", Value: userID},
		{Key: "tweets", Value: []*tweet.Tweet{}},
		{Key: "followers", Value: []string{}},
		{Key: "followed", Value: []string{}},
		{Key: "timeline", Value: []*tweet.Tweet{}},
	})
	if err != nil {
		return nil, fmt.Errorf("ERROR: %v\n", err.Error())
	}

	return user, nil
}

func (s *Storage) GetUser(ctx context.Context, userID string) (*tweet.User, error) {
	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{
					{"user_id", userID},
				},
			},
		},
	}

	var result bson.M

	if err := s.db.Collection("users").FindOne(ctx, filter).Decode(&result); err != nil {
		return nil, err
	}

	var user *tweet.User

	bsonBytes, _ := bson.Marshal(result)
	bson.Unmarshal(bsonBytes, &user)

	return user, nil
}

func (s *Storage) CreateTweet(ctx context.Context, userID string, data string) (*tweet.Tweet, error) {
	tweet := &tweet.Tweet{
		ID:        uuid.NewString(),
		Data:      data,
		UserID:    userID,
		TimeStamp: time.Now(),
	}

	filter := bson.D{{"user_id", userID}}
	update := bson.D{{"$push",
		bson.D{
			{"tweets", tweet},
		},
	}}

	var collectionName = "users"
	collectionTweets := s.db.Collection(collectionName)

	_, err := collectionTweets.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("ERROR: %v\n", err.Error())
	}

	return tweet, nil
}

func (s *Storage) Follow(ctx context.Context, user *tweet.User, userToFollow *tweet.User) error {
	var collectionName = "users"
	collectionTweets := s.db.Collection(collectionName)

	filter := bson.D{{"user_id", user.UserID}}
	update := bson.D{{"$push",
		bson.D{
			{"followed", userToFollow.UserID},
		},
	}}

	_, err := collectionTweets.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ERROR: %v\n", err.Error())
	}

	filter = bson.D{{"user_id", userToFollow.UserID}}
	update = bson.D{{"$push",
		bson.D{
			{"followers", user.UserID},
		},
	}}

	_, err = collectionTweets.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ERROR: %v\n", err.Error())
	}

	return nil
}

func (s *Storage) UpdateTimeLIne(ctx context.Context, user *tweet.User, tweet *tweet.Tweet) error {
	var collectionName = "users"
	collectionTweets := s.db.Collection(collectionName)

	filter := bson.D{{"user_id", user.UserID}}
	update := bson.D{{"$push",
		bson.D{
			{"timeline", tweet},
		},
	}}

	_, err := collectionTweets.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ERROR: %v\n", err.Error())
	}

	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
