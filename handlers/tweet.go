package handlers

import (
	"challenge/internal/tweet"
	"context"
	"encoding/json"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

type tweetService interface {
	CreateTweet(ctx context.Context, userID string, tweet *tweet.Request) error
	Follow(ctx context.Context, id string, follow string) error
	ViewTimeline(ctx context.Context, userID string) ([]*tweet.Tweet, error)
}

type HandlerTweet struct {
	service tweetService
}

type Error struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

func NewTweetHandler(tweetService tweetService) *HandlerTweet {
	return &HandlerTweet{service: tweetService}
}

func (h *HandlerTweet) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		handleError(generateErrorResponse("header user id is not present", http.StatusBadRequest), w)
		return
	}

	var tweet *tweet.Request
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		handleError(generateErrorResponse("invalid body", http.StatusBadRequest), w)
		return
	}

	err := h.service.CreateTweet(r.Context(), userID, tweet)
	if err != nil {
		handleError(generateErrorResponse(err.Error(), http.StatusInternalServerError), w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (h *HandlerTweet) Follow(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		handleError(generateErrorResponse("header user id is not present", http.StatusBadRequest), w)
		return
	}

	userIDToFollow := chi.URLParam(r, "userID")
	if strings.TrimSpace(userID) == "" {
		handleError(generateErrorResponse("user id to follow is invalid", http.StatusBadRequest), w)
		return
	}

	if userID == userIDToFollow {
		handleError(generateErrorResponse("user ID to follow must be other", http.StatusBadRequest), w)
	}

	err := h.service.Follow(r.Context(), userID, userIDToFollow)
	if err != nil {
		handleError(generateErrorResponse(err.Error(), http.StatusInternalServerError), w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

func (h *HandlerTweet) ViewTimeline(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		handleError(generateErrorResponse("header user id is not present", http.StatusBadRequest), w)
		return
	}

	timeLine, err := h.service.ViewTimeline(r.Context(), userID)
	if err != nil {
		handleError(generateErrorResponse(err.Error(), http.StatusInternalServerError), w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeLine)
}

func generateErrorResponse(msg string, status int) *Error {
	return &Error{
		Error:  msg,
		Status: status,
	}
}

func handleError(error *Error, w http.ResponseWriter) {
	w.WriteHeader(error.Status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(generateErrorResponse(error.Error, error.Status))
}
