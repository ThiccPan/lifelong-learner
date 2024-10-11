package main

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/thiccpan/lifelong-learner/internal/common/auth"
	"google.golang.org/api/iterator"
)

type UserServer struct {
	db *firestore.Client
}

func NewUserServer(db *firestore.Client) *UserServer {
	return &UserServer{
		db: db,
	}
}

type GetCurrentUserRequest struct {
}

func (uh *UserServer) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCotext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(err.Error()))
	}
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

	data := res.Data()
	dataInBytes, _ := json.Marshal(data)
	w.Write([]byte(dataInBytes))
}
