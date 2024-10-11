package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	appAuth "github.com/thiccpan/lifelong-learner/internal/common/auth"
	"google.golang.org/api/option"
)

const (
	PORT string = "8080"
)

func main() {
	ctx := context.Background()
	db := firebaseSetup()
	store, err := db.Firestore(ctx)
	if err != nil {
		panic("err init db")
	}
	// authService, err := db.Auth(ctx)
	// if err != nil {
	// 	panic("err init firebase auth service")
	// }

	userHandler := &UserServer{
		db:   store,
	}
	rootRouter := chi.NewRouter()

	usersRouter := chi.NewRouter()
	setMiddlewares(usersRouter)
	usersMux := HandlerFromMux(
		&ServerInterfaceWrapper{
			Handler: userHandler,
		}, usersRouter)
	rootRouter.Mount("/api", usersMux)

	slog.Info("starting http server at port: " + PORT)
	http.ListenAndServe(":"+PORT, rootRouter)
}

func setMiddlewares(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	addCorsMiddleware(r)
	addAuthMiddleware(r)

	r.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)

}

func addAuthMiddleware(r *chi.Mux) {
	authService, err := firebaseSetup().Auth(context.Background())
	if err != nil {
		slog.Error("failed to init firebase auth")
		panic(1)
	}
	auth := appAuth.FirebaseAuth{AuthClient: authService}
	r.Use(auth.FirebaseAuthMiddleware)
}

func addCorsMiddleware(r *chi.Mux) {
	origins := []string{"*"}
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"link"},
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)
}

func firebaseSetup() *firebase.App {
	opt := option.WithCredentialsFile("secrets/firebase-key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	return app
}
