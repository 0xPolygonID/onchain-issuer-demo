package common

import (
	"os"
)

var (
	ExternalServerHost string
	InternalServerPort string
	MongoDBHost        string
)

func init() {
	ExternalServerHost = os.Getenv("EXTERNAL_SERVER_HOST")
	if ExternalServerHost == "" {
		panic("SERVER_HOST env variable is not set")
	}
	InternalServerPort = os.Getenv("INTERNAL_SERVER_PORT")
	if InternalServerPort == "" {
		InternalServerPort = "3333"
	}
	MongoDBHost = os.Getenv("MONGODB_HOST")
	if MongoDBHost == "" {
		MongoDBHost = "mongodb://localhost:27017/credentials/?timeoutMS=5000"
	}
}
