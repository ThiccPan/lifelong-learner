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

	// firebase setup
	firebaseService := firebaseSetup()
	store, err := firebaseService.Firestore(ctx)
	if err != nil {
		panic("err init db")
	}

	// cors and firebase auth setup
	corsMiddleware := newCorsMiddleware()
	firebaseAuthService := newAuthMiddleware()

	// handler for user domain setup
	userHandler := &UserServer{
		db:   store,
		auth: firebaseAuthService,
	}

	// declaring root router
	rootRouter := chi.NewRouter()
	// declaring users router that implements handler interface generated by the specs
	usersRouter := chi.NewRouter()

	// applying the middleware stack to the users router
	setMiddlewares(usersRouter,
		corsMiddleware.Handler,
		firebaseAuthService.FirebaseAuthMiddleware)

	// implementing handler to the defined router
	usersMux := HandlerFromMux(
		&ServerInterfaceWrapper{
			Handler: userHandler,
		}, usersRouter)
	rootRouter.Mount("/api", usersMux)

	slog.Info("starting http server at port: " + PORT)
	http.ListenAndServe(":"+PORT, rootRouter)
}

func firebaseSetup() *firebase.App {
	opt := option.WithCredentialsFile("secrets/firebase-key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	return app
}

func setMiddlewares(r *chi.Mux, opts ...func(http.Handler) http.Handler) {
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	for _, handler := range opts {
		r.Use(handler)
	}

	r.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)

}

func newAuthMiddleware() *appAuth.FirebaseAuth {
	authService, err := firebaseSetup().Auth(context.Background())
	if err != nil {
		slog.Error("failed to init firebase auth")
		panic(1)
	}
	return &appAuth.FirebaseAuth{AuthClient: authService}
}

func newCorsMiddleware() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Control-Allow-Headers", "Origin", "Accept", "X-Requested-With", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"link"},
		MaxAge:           300,
	})
}
