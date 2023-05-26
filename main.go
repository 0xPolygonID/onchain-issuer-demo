package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/iden3/issuer-on-chain-backend/common"
	"github.com/iden3/issuer-on-chain-backend/handlers"
	"github.com/iden3/issuer-on-chain-backend/repository"
	"github.com/iden3/issuer-on-chain-backend/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cr, err := initRepository()
	if err != nil {
		log.Fatal("failed connect to mongodb:", err)
	}
	onchainService := services.OnChain{
		CredentialRepository: cr,
	}
	h := handlers.Handlers{
		CredentialService: onchainService,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Minute))

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/identities/{identifier}/claims", h.CreateClaim)
			r.Post("/agent", h.Handle)
			// r.Get("/identities/{identifier}/claims", handlers.GetClaimsList)
			// r.Get("/identities/{identifier}/claims/{claimId}", handlers.GetClaim)
		})
	})

	http.ListenAndServe(":3333", r)
}

func initRepository() (*repository.CredentialRepository, error) {
	db, err := mongo.NewClient(options.Client().ApplyURI(common.MongoDBHost))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = db.Connect(ctx)
	if err != nil {
		log.Fatal("Context error, mongoDB:", err)
	}

	return repository.NewCredentialRepository(db.Database("master"))
}
