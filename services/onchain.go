package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/issuer-on-chain-backend/repository"
)

type OnChain struct {
	CredentialRepository *repository.CredentialRepository
	// chainContract
}

type CredentialRequest struct {
	CredentialSchema      string          `json:"credentialSchema"`
	Type                  string          `json:"type"`
	CredentialSubject     json.RawMessage `json:"credentialSubject"`
	Expiration            int64           `json:"expiration,omitempty"`
	Version               uint32          `json:"version,omitempty"`
	RevNonce              *uint64         `json:"revNonce,omitempty"`
	SubjectPosition       string          `json:"subjectPosition,omitempty"`
	MerklizedRootPosition string          `json:"merklizedRootPosition,omitempty"`
}

// CreateClaimOnChain create onchain vc to issuer
func (oc *OnChain) CreateClaimOnChain(
	ctx context.Context,
	issuer string,
	credentialReq CredentialRequest,
) (string, error) {
	schemaBytes, err := loadSchema(context.Background(), credentialReq.CredentialSchema)
	if err != nil {
		return "", err
	}
	var schema jsonSuite.Schema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return "", err
	}

	w3cCred, err := CreateW3CCredential(
		schema, issuer, credentialReq)
	if err != nil {
		return "", err
	}

	coreClaim, err := BuildCoreClaim(
		schema, schemaBytes, w3cCred, credentialReq,
		w3cCred.CredentialStatus.(verifiable.CredentialStatus).RevocationNonce,
		credentialReq.Version)
	if err != nil {
		return "", err
	}

	// TODO: write to smart contract
	id, err := oc.CredentialRepository.Create(ctx, w3cCred)
	if err != nil {
		return "", err
	}

	mustPrintCoreClain(coreClaim)
	mustPritnVC(w3cCred)

	return id, nil
}

func loadSchema(ctx context.Context, URL string) ([]byte, error) {
	resp, err := http.DefaultClient.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func mustPrintCoreClain(coreClaim *core.Claim) {
	h, err := coreClaim.Hex()
	if err != nil {
		panic(err)
	}
	fmt.Println("core claim:", h)
}

func mustPritnVC(w3cCred verifiable.W3CCredential) {
	raw, err := json.Marshal(w3cCred)
	if err != nil {
		panic(err)
	}
	fmt.Println("vc:", string(raw))
}
