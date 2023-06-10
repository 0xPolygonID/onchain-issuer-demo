package services

import (
	"context"
	"fmt"

	// neet to update go-schema-processor to core v2
	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	proc "github.com/iden3/go-schema-processor/processor"
	jsonproc "github.com/iden3/go-schema-processor/processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
)

func BuildCoreClaim(
	schemaSuite jsonSuite.Schema,
	schemaBytes []byte,
	vc verifiable.W3CCredential,
	credentialReq CredentialRequest,
	revocationNonce uint64,
	version uint32,
) (*core.Claim, error) {

	jsonProcessor := jsonproc.New(
		proc.WithParser(jsonSuite.Parser{}),
	)
	opts := proc.CoreClaimOptions{
		RevNonce:        revocationNonce,
		Version:         version,
		SubjectPosition: credentialReq.SubjectPosition,
		MerklizedRootPosition: defineMerklizedRootPosition(
			schemaSuite.Metadata,
			credentialReq.MerklizedRootPosition,
		),
		Updatable: false,
	}

	credentialType := fmt.Sprintf("%s#%s", schemaSuite.Metadata.Uris["jsonLdContext"], credentialReq.Type)
	fmt.Println("credentialType str", credentialType)

	coreClaim, err := jsonProcessor.ParseClaim(
		context.Background(),
		vc,
		credentialType,
		schemaBytes,
		&opts,
	)
	if err != nil {
		return nil, err
	}
	return coreClaim, nil
}

func defineMerklizedRootPosition(metadata *jsonSuite.SchemaMetadata, position string) string {
	if metadata != nil && metadata.Serialization != nil {
		return ""
	}
	if position != "" {
		return position
	}
	return verifiable.CredentialMerklizedRootPositionIndex
}
