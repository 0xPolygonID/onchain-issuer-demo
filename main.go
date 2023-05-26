package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0xPolygonID/issuer-on-chain-backend/common"
	"github.com/0xPolygonID/issuer-on-chain-backend/handlers"
	"github.com/0xPolygonID/issuer-on-chain-backend/repository"
	"github.com/0xPolygonID/issuer-on-chain-backend/services"
	"github.com/pkg/errors"
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
			// r.Get("/identities/{identifier}/claims", handlers.GetClaimsList)
			// r.Get("/identities/{identifier}/claims/{claimId}", handlers.GetClaim)
		})
	})

	http.ListenAndServe(":3333", r)
}

func initRepository() (*repository.CredentialRepository, error) {
	fmt.Println("Connecting to MongoDB: ", common.MongoDBHost)
	opts := options.Client().ApplyURI(common.MongoDBHost)
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mongo")
	}
	if err = client.Ping(context.Background(), nil); err != nil {
		return nil, errors.Wrap(err, "failed to ping mongo")
	}
	rep, err := repository.NewCredentialRepository(client.Database("credentials"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create credential repository")
	}
	return rep, nil
}
