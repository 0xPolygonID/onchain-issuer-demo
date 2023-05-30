package repository

import (
	"fmt"
	"reflect"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

type Data struct {
	Type string
}

func (d *Data) ProofType() verifiable.ProofType {
	return ""
}

func (d *Data) GetCoreClaim() (*core.Claim, error) {
	return nil, nil
}

type CredentialProofsCodec struct{}

func (c *CredentialProofsCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	fmt.Println("can set?", val.CanSet())

	vr.Skip()

	// topLevelArr, err := vr.ReadArray()
	// if err != nil {
	// 	return err
	// }
	// index0, err := topLevelArr.ReadValue()
	// if err != nil {
	// 	return err
	// }
	// mtpProof, err := index0.ReadDocument()
	// if err != nil {
	// 	return err
	// }
	// key, typeValue, err := mtpProof.ReadElement()
	// if err != nil {
	// 	return err
	// }
	// if key != "type" {
	// 	return fmt.Errorf("expected type, got %s", key)
	// }
	// typeStr, err := typeValue.ReadString()
	// if err != nil {
	// 	return err
	// }

	// d := verifiable.CredentialProofs{
	// 	&Data{
	// 		Type: typeStr,
	// 	},
	// }
	// srcPtr := reflect.ValueOf(d)
	// val.Set(srcPtr)

	return nil
}
