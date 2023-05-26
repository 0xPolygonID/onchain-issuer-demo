package repository

import (
	"context"
	"errors"

	"github.com/iden3/go-schema-processor/verifiable"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CredentialRepository struct {
	db   *mongo.Database
	coll *mongo.Collection
}

func NewCredentialRepository(db *mongo.Database) (*CredentialRepository, error) {
	err := db.CreateCollection(
		context.Background(),
		"credentials",
		options.CreateCollection().SetCollation(
			&options.Collation{
				Locale:   "en",
				Strength: 2,
			},
		),
	)
	if err != nil {
		var comErr mongo.CommandError
		if errors.As(err, &comErr) && comErr.Code == 48 {
			// collection already exists
		} else {
			return nil, err
		}
	}

	collection := db.Collection("credentials")
	_, err = collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   "id",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{
					Key:   "issuer",
					Value: 1,
				},
			},
		},
		{
			Keys: bson.D{
				{
					Key:   "credentialSubject.id",
					Value: 1,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &CredentialRepository{
		db:   db,
		coll: collection,
	}, nil
}

func (cs *CredentialRepository) Create(
	ctx context.Context,
	vc verifiable.W3CCredential,
) (string, error) {
	res, err := cs.coll.InsertOne(ctx, vc)
	if err != nil {
		return "", err
	}
	id := res.InsertedID.(primitive.ObjectID)
	return id.Hex(), nil
}
