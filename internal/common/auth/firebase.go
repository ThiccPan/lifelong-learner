package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

var (
	ErrorNotAuthenticatedUser = errors.New("user is not authenticated")
	ErrorInvalidToken         = errors.New("invalid auth token")
)

type FirebaseAuth struct {
	AuthClient *auth.Client
}

func (fa *FirebaseAuth) FirebaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get auth token string from the request header
		authTokenString := r.Header.Get("Authorization")
		// fmt.Println(authTokenString)
		if authTokenString == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(ErrorNotAuthenticatedUser.Error()))
			return
		}

		// check if auth token string is a valid bearer token
		if len(authTokenString) < 7 && strings.ToLower(authTokenString[:6]) != "bearer" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(ErrorInvalidToken.Error()))
			return
		}

		if strings.ToLower(authTokenString[:6]) == "bearer" {
			authTokenString = authTokenString[7:]
		}

		// verify using firebase verification method
		authToken, err := fa.AuthClient.VerifyIDTokenAndCheckRevoked(ctx, authTokenString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		ctx = context.WithValue(ctx, userContextKey, User{
			UUID:  authToken.UID,
			Email: authToken.Claims["email"].(string),
		})

		fmt.Println("context:", userContextKey, ctx.Value(userContextKey))

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type ContextKey int

const (
	userContextKey ContextKey = iota
)

type User struct {
	UUID        string
	Email       string
	Role        string
	DisplayName string
}

var (
	ErrorNoUserInContext = errors.New("no user in context err")
)

func UserFromCotext(ctx context.Context) (*User, error) {
	fmt.Println("context:", userContextKey, ctx.Value(userContextKey))
	user, ok := ctx.Value(userContextKey).(User)
	if !ok {
		return nil, ErrorNoUserInContext
	}
	return &user, nil
}
