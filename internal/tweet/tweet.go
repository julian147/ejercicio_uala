package tweet

import (
	"context"
	"fmt"
	"time"
)

//go:generate mockgen -source=tweet.go -destination=tweet_mock.go -package=tweet

type Tweet struct {
	ID        string
	Data      string
	UserID    string
	TimeStamp time.Time
}

type Request struct {
	Data string `json:"data"`
}

type User struct {
	UserID    string   `bson:"user_id"`
	Followed  []string `bson:"followed"`
	Followers []string `bson:"followers"`
	Tweets    []*Tweet `bson:"tweets"`
	Timeline  []*Tweet `bson:"timeline"`
}

type Storage interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	CreateTweet(ctx context.Context, userID string, data string) (*Tweet, error)
	CreateUser(ctx context.Context, userID string) (*User, error)
	Follow(ctx context.Context, user *User, userToFollow *User) error
	UpdateTimeLIne(ctx context.Context, user *User, tweet *Tweet) error
}

type Manager struct {
	storage Storage
}

func New(storage Storage) *Manager {
	return &Manager{
		storage: storage,
	}
}

func (s *Manager) CreateTweet(ctx context.Context, userID string, tweetRequest *Request) error {
	var user *User

	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		if err.Error() != "mongo: no documents in result" {
			return fmt.Errorf("getting users")
		} else {
			user, err = s.storage.CreateUser(ctx, userID)
			if err != nil {
				return fmt.Errorf("creating user")
			}
		}

	}

	tweet, err := s.storage.CreateTweet(ctx, userID, tweetRequest.Data)
	if err != nil {
		return fmt.Errorf("creating tweet")
	}

	childContext, _ := context.WithTimeout(context.Background(), 5*time.Second)

	go s.updateTimeline(childContext, user, tweet)

	return err
}

func (s *Manager) GetUser(ctx context.Context, userID string) (*User, error) {
	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return user, nil
}

func (s *Manager) Follow(ctx context.Context, userID string, userIDToFollow string) error {
	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	for _, userFollowed := range user.Followed {
		if userFollowed == userIDToFollow {
			return fmt.Errorf("you are follow this user")
		}
	}

	userToFollow, err := s.storage.GetUser(ctx, userIDToFollow)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	err = s.storage.Follow(ctx, user, userToFollow)
	if err != nil {
		return fmt.Errorf("following user: %w", err)
	}

	return err
}

func (s *Manager) ViewTimeline(ctx context.Context, userID string) ([]*Tweet, error) {
	user, err := s.storage.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	return user.Timeline, nil
}

func (s *Manager) updateTimeline(ctx context.Context, user *User, tweet *Tweet) {
	for _, follower := range user.Followers {
		user, err := s.storage.GetUser(ctx, follower)
		if err != nil {
			return
		}
		err = s.storage.UpdateTimeLIne(ctx, user, tweet)
		if err != nil {
			return
		}
	}

	return
}
