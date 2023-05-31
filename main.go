package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/0xPolygonID/onchain-issuer-demo/common"
	"github.com/0xPolygonID/onchain-issuer-demo/handlers"
	"github.com/0xPolygonID/onchain-issuer-demo/repository"
	"github.com/0xPolygonID/onchain-issuer-demo/services"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-jwz"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm"
	"github.com/iden3/iden3comm/packers"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
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
	pkg, err := initPakcer()
	if err != nil {
		log.Fatal("failed init packer:", err)
	}
	onchainService := services.OnChain{
		CredentialRepository: cr,
	}

	h := handlers.Handlers{
		CredentialService: onchainService,
		Packager:          pkg,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * time.Minute))

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/identities/{identifier}/claims", h.CreateClaim)
			r.Get("/identities/{identifier}/claims", h.GetUserVCs)
			r.Get("/identities/{identifier}/claims/{claimId}", h.GetUserVCByID)
			r.Post("/agent", h.Agent)
		})
	})

	http.ListenAndServe(":3333", r)
}

func initRepository() (*repository.CredentialRepository, error) {
	tM := reflect.TypeOf(bson.M{})
	reg := bson.NewRegistryBuilder().
		RegisterTypeDecoder(reflect.TypeOf(verifiable.CredentialProofs{}), &repository.CredentialProofsCodec{}).
		RegisterTypeMapEntry(bsontype.EmbeddedDocument, tM).Build()

	fmt.Println("Connecting to MongoDB: ", common.MongoDBHost)
	opts := options.Client().ApplyURI(common.MongoDBHost).SetRegistry(reg)
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

func initPakcer() (*iden3comm.PackageManager, error) {
	stateContracts := map[string]*abi.State{}
	for chainPrefix, resolverSetting := range common.ResolverSettings {
		client, err := ethclient.Dial(resolverSetting.NetworkURL)
		if err != nil {
			return nil, err
		}
		add := ethcomm.HexToAddress(resolverSetting.ContractState)
		stateContract, err := abi.NewState(add, client)
		if err != nil {
			return nil, err
		}
		stateContracts[chainPrefix] = stateContract
	}

	authV2VerificationKey, err := os.ReadFile(common.AuthV2VerificationKeyPath)
	if err != nil {
		return nil, err
	}
	unpakOpts := map[jwz.ProvingMethodAlg]packers.VerificationParams{
		jwz.AuthV2Groth16Alg: {
			Key:            authV2VerificationKey,
			VerificationFn: services.StateVerificationHandler(stateContracts),
		},
	}
	zkpPacker := packers.NewZKPPacker(
		nil,
		unpakOpts,
	)
	packer := iden3comm.NewPackageManager()
	packer.RegisterPackers(
		zkpPacker,
		&packers.PlainMessagePacker{},
	)

	return packer, nil
}
