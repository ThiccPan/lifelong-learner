package main

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/thiccpan/lifelong-learner/internal/common/auth"
	"google.golang.org/api/iterator"
)

type UserServer struct {
	db   *firestore.Client
	auth *auth.FirebaseAuth
}

func NewUserServer(db *firestore.Client) *UserServer {
	return &UserServer{
		db: db,
	}
}

type GetCurrentUserRequest struct {
}

type GetCurrentUserResponse struct {
	UserData User   `json:"user_data"`
	Token    string `json:"token"`
}

func (uh *UserServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// get user email from auth
	user, err := auth.UserFromCotext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
	}
	// fetch user data for authorization management using token
	resQuery := uh.db.
		Collection("users").
		Where("email", "==", user.Email).
		Limit(1).
		Documents(r.Context())

	res, err := resQuery.Next()
	if err == iterator.Done {
		http.Error(w, "user data not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// formatting the data
	data := res.Data()
	userRole := data["role"].(string)
	userName := data["displayName"].(string)
	userBalance := data["balance"].(int64)

	// updating old token to include user data for authorization purpose
	if err := uh.auth.CreateCustomToken(
		user.UUID,
		userRole,
		userName,
		int(userBalance),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"message": "success",
	})
}

func (uh *UserServer) Test(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]any{
		"message": "success",
	})
}
