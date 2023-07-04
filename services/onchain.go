package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/0xPolygonID/onchain-issuer-demo/common"
	"github.com/0xPolygonID/onchain-issuer-demo/pkg/blockchain"
	"github.com/0xPolygonID/onchain-issuer-demo/repository"
	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
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

	contractAddress, err := buildContractIDFromDID(issuer)
	if err != nil {
		return "", err
	}

	rs, ok := common.OnChainIssuerSettings[contractAddress]
	if !ok {
		return "", fmt.Errorf("resolver settings for chain %s not found", contractAddress)
	}

	w3cCred, err := CreateW3CCredential(
		schema, issuer, credentialReq, rs.ChainID)
	if err != nil {
		return "", err
	}
	mustPritnVC(w3cCred)

	coreClaim, err := BuildCoreClaim(
		schema, schemaBytes, w3cCred, credentialReq,
		w3cCred.CredentialStatus.(verifiable.CredentialStatus).RevocationNonce,
		credentialReq.Version)
	if err != nil {
		return "", err
	}
	mustPrintCoreClain(coreClaim)

	hi, err := coreClaim.HIndex()
	if err != nil {
		return "", err
	}
	hv, err := coreClaim.HValue()
	if err != nil {
		return "", err
	}
	mtpProof, err := blockchain.ProcessOnChainClaim(
		rs.NetworkURL,
		contractAddress,
		rs.ContractOwner,
		hi,
		hv,
	)
	if err != nil {
		return "", err
	}

	mtpProof.IssuerData.ID = issuer
	mtpProof.CoreClaim, _ = coreClaim.Hex()
	w3cCred.Proof = verifiable.CredentialProofs{
		&mtpProof,
	}

	id, err := oc.CredentialRepository.Create(ctx, w3cCred)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (oc *OnChain) GetUsersVCs(
	ctx context.Context,
	issuer string,
	subject string,
	schemaType string,
) ([]verifiable.W3CCredential, error) {
	return oc.CredentialRepository.GetUserVCs(ctx, issuer, subject, schemaType)
}

func (oc *OnChain) GetUserVCByID(
	ctx context.Context,
	issuer string,
	credentialID string,
) (verifiable.W3CCredential, error) {
	return oc.CredentialRepository.GetVCByID(ctx, issuer, credentialID)
}

func (oc *OnChain) IsRevokedVC(
	ctx context.Context,
	issuer string,
	nonce uint64,
) (bool, error) {
	contractAddress, err := buildContractIDFromDID(issuer)
	if err != nil {
		return false, err
	}

	rs, ok := common.OnChainIssuerSettings[contractAddress]
	if !ok {
		return false, fmt.Errorf("resolver settings for chain %s not found", contractAddress)
	}

	return blockchain.IsRevokedClaim(
		rs.NetworkURL,
		contractAddress,
		nonce,
	)
}

func (oc *OnChain) RevokeVC(
	ctx context.Context,
	issuer string,
	nonce uint64,
) error {
	contractAddress, err := buildContractIDFromDID(issuer)
	if err != nil {
		return err
	}

	rs, ok := common.OnChainIssuerSettings[contractAddress]
	if !ok {
		return fmt.Errorf("resolver settings for chain %s not found", contractAddress)
	}

	return blockchain.RevokeOnChainClaim(
		rs.NetworkURL,
		contractAddress,
		rs.ContractOwner,
		nonce,
	)
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
