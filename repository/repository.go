package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/iden3/go-schema-processor/verifiable"
	"go.mongodb.org/mongo-driver/bson"
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
					Key:   "credentialsubject.id",
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
	model, err := NewCredentailModelFromW3C(vc)
	if err != nil {
		return "", err
	}
	_, err = cs.coll.InsertOne(ctx, model)
	if err != nil {
		return "", err
	}
	return extractCredentialID(vc), nil
}

func (cs *CredentialRepository) GetUserVCs(
	ctx context.Context,
	issuer string,
	user string,
	schemaType string,
) ([]verifiable.W3CCredential, error) {
	filter := bson.M{"issuer": issuer}
	if user != "" {
		filter["credentialsubject.id"] = user
	}
	if schemaType != "" {
		filter["credentialsubject.type"] = schemaType
	}
	cursor, err := cs.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var models []credentialModel
	if err = cursor.All(ctx, &models); err != nil {
		return nil, err
	}

	vcs := make([]verifiable.W3CCredential, len(models))
	for i, model := range models {
		vc, err := model.ToW3C()
		if err != nil {
			return nil, err
		}
		vcs[i] = vc
	}

	return vcs, nil
}

func (cs *CredentialRepository) GetVCByID(
	ctx context.Context,
	issuer string,
	credentialID string,
) (verifiable.W3CCredential, error) {
	filter := bson.M{"issuer": issuer, "id": bson.M{"$regex": credentialID}}
	res := cs.coll.FindOne(ctx, filter)
	if res.Err() != nil {
		return verifiable.W3CCredential{}, res.Err()
	}
	var model credentialModel
	if err := res.Decode(&model); err != nil {
		return verifiable.W3CCredential{}, err
	}
	return model.ToW3C()
}

func extractCredentialID(vc verifiable.W3CCredential) string {
	parts := strings.Split(vc.ID, "/")
	return parts[len(parts)-1]
}
