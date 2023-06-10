package services

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xPolygonID/onchain-issuer-demo/common"
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
)

var (
	defaultW3Cctx = [2]string{
		verifiable.JSONLDSchemaW3CCredential2018,
		"https://raw.githubusercontent.com/iden3/claim-schema-vocab/cbade52faccea8c386bab0129c0ffffa64393849/core/jsonld/iden3proofs.jsonld",
	}
	defaultW3CCredentialType = [1]string{
		verifiable.TypeW3CVerifiableCredential,
	}
)

func CreateW3CCredential(
	schema jsonSuite.Schema,
	issuer string,
	req CredentialRequest,
	chainID int,
) (verifiable.W3CCredential, error) {
	expirationTime := time.Unix(req.Expiration, 0)
	issuanceDate := time.Now()

	credentialSubject := map[string]interface{}{
		"type": req.Type,
	}
	err := json.Unmarshal(req.CredentialSubject, &credentialSubject)
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	if err := fillCredentialSubject(credentialSubject); err != nil {
		return verifiable.W3CCredential{}, err
	}

	cs, err := buildCredentialStatus(issuer, chainID)
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	cred := verifiable.W3CCredential{
		ID:                newCredentialID(issuer),
		Context:           append(defaultW3Cctx[:], schema.Metadata.Uris["jsonLdContext"].(string)),
		Type:              append(defaultW3CCredentialType[:], req.Type),
		IssuanceDate:      &issuanceDate,
		CredentialSubject: credentialSubject,
		Issuer:            issuer,
		CredentialStatus:  cs,
		CredentialSchema: verifiable.CredentialSchema{
			ID:   req.CredentialSchema,
			Type: verifiable.JSONSchemaValidator2018,
		},
	}
	if req.Expiration != 0 {
		cred.Expiration = &expirationTime
	}
	return cred, nil
}

func fillCredentialSubject(credentialSubject map[string]interface{}) error {
	idSubject, ok := credentialSubject["id"].(string)
	if ok {

		did, err := w3c.ParseDID(idSubject)
		if err != nil {
			return err
		}

		credentialSubject["id"] = did.String()
	}
	return nil
}

func buildCredentialStatus(issuer string, chainID int) (verifiable.CredentialStatus, error) {
	cid, err := buildContractIDFromDID(issuer)
	if err != nil {
		return verifiable.CredentialStatus{}, err
	}
	uintNonce, err := RandInt64()
	if err != nil {
		return verifiable.CredentialStatus{}, err
	}
	return verifiable.CredentialStatus{
		ID: fmt.Sprintf(
			"%s/credentialStatus?revocationNonce=%d&contractAddress=%d:%s", issuer, uintNonce, chainID, cid,
		),
		RevocationNonce: uintNonce,
		Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
	}, nil
}

func RandInt64() (uint64, error) {
	var buf [8]byte
	// TODO: this was changed because revocation nonce is cut in dart / js if number is too big
	_, err := rand.Read(buf[:4]) // was rand.Read(buf[:])

	return binary.LittleEndian.Uint64(buf[:]), err
}

func newCredentialID(issuer string) string {
	return fmt.Sprintf(
		"%s/api/v1/identities/%s/claims/%s",
		common.ExternalServerHost, issuer, uuid.New().String())
}

func buildContractIDFromDID(didStr string) (string, error) {
	did, err := w3c.ParseDID(didStr)
	if err != nil {
		return "", err
	}
	id, err := core.IDFromDID(*did)
	if err != nil {
		return "", err
	}
	ca, err := core.EthAddressFromID(id)
	if err != nil {
		return "", err
	}
	return ethcomm.BytesToAddress(ca[:]).Hex(), nil
}
