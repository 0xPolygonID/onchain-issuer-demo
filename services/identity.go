package services

import (
	"context"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/issuer-on-chain-backend/repository"
)

// IdentityService service
type IdentityService struct {
	CredentialRepository *repository.CredentialRepository
}

// Exists get latest identity state by identifier
func (i *IdentityService) Exists(ctx context.Context,
	identifier *core.DID) (bool, error) {
	// check that identity exists in the db
	identity, err := i.CredentialRepository.GetIdentityByID(ctx, identifier)
	if err != nil {
		return false, err
	}
	if identity == nil {
		return false, nil
	}
	return true, nil
}
