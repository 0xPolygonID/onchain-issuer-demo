package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"
	"github.com/iden3/issuer-on-chain-backend/repository"
	"github.com/iden3/issuer-on-chain-backend/services"
)

type Handlers struct {
	packager             *iden3comm.PackageManager
	CredentialService    services.OnChain
	CredentialRepository *repository.CredentialRepository
	identityService      services.IdentityService
}

func (h *Handlers) CreateClaim(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	credentialReq := services.CredentialRequest{}
	if err := json.NewDecoder(r.Body).Decode(&credentialReq); err != nil {
		// TODO: move errors to one plase
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recordID, err := h.CredentialService.CreateClaimOnChain(r.Context(), issuer, credentialReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": recordID})
}

func (h *Handlers) Handle(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1000*1000)
	envelope, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't bind request to protocol message", http.StatusBadRequest)
		return
	}

	basicMessage, err := h.packager.UnpackWithType(packers.MediaTypeZKPMessage, envelope)
	if err != nil {
		http.Error(w, "failed unpack protocol message", http.StatusBadRequest)
		return
	}

	if basicMessage.ID == "" {
		http.Error(w, "empty 'id' field", http.StatusBadRequest)
		return
	}

	if basicMessage.To == "" {
		http.Error(w, "empty 'to' field", http.StatusBadRequest)
		return
	}
	var to *core.DID
	to, err = core.ParseDID(basicMessage.To)
	if err != nil {
		http.Error(w, "no target identity in protocol message", http.StatusBadRequest)
		return
	}

	exists, err := h.identityService.Exists(r.Context(), to)
	if err != nil {
		http.Error(w, "can't get target identity", http.StatusBadRequest)
		return
	}
	if !exists {
		http.Error(w, "target identity is not managed by agent", http.StatusBadRequest)
		return
	}
	var (
		resp           []byte
		httpStatusCode = http.StatusOK
	)

	switch basicMessage.Type {
	case protocol.RevocationStatusRequestMessageType:
		http.Error(w, "failed handling revocation status request", http.StatusBadRequest)
		return
	case protocol.CredentialFetchRequestMessageType:
		resp, err = h.handleCredentialFetchRequest(r.Context(), basicMessage)
		if err != nil {
			http.Error(w, "failed handling credential fetch request", http.StatusBadRequest)
			return
		}
	}

	_, err = core.ParseDID(basicMessage.From)
	if err != nil {
		http.Error(w, "failed get sender from request", http.StatusBadRequest)
		return
	}

	var respBytes []byte
	if len(resp) > 0 {
		respBytes, err = h.packager.Pack(packers.MediaTypePlainMessage, resp, packers.PlainPackerParams{})
		if err != nil {
			http.Error(w, "failed create jwz token", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	_, err = w.Write(respBytes)

}

func (h *Handlers) handleCredentialFetchRequest(ctx context.Context, basicMessage *iden3comm.BasicMessage) ([]byte, error) {
	if basicMessage.To == "" {
		return nil, errors.New("failed request. empty 'to' field")
	}

	if basicMessage.From == "" {
		return nil, errors.New("failed request. empty 'from' field")
	}

	fetchRequestBody := &protocol.CredentialFetchRequestMessageBody{}
	err := json.Unmarshal(basicMessage.Body, fetchRequestBody)
	if err != nil {
		return nil, fmt.Errorf("invalid credential fetch request body: %w", err)
	}

	var issuerDID *core.DID
	issuerDID, err = core.ParseDID(basicMessage.To)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer id in base message: %w", err)
	}

	// var userDID *core.DID
	// userDID, err = core.ParseDID(basicMessage.From)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid user id in base message: %w", err)
	// }

	var claimID uuid.UUID
	claimID, err = uuid.Parse(fetchRequestBody.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid claim id in fetch request body: %w", err)
	}

	cred, err := h.CredentialRepository.GetCredentialById(ctx, issuerDID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed get claim by claimID: %w", err)
	}

	// if claim.OtherIdentifier != userDID.String() {
	// 	return nil, errors.New("claim doesn't relate to sender")
	// }

	if err != nil {
		return nil, fmt.Errorf("failed convert claim: %w", err)
	}

	resp, err := json.Marshal(&protocol.CredentialIssuanceMessage{
		ID:       uuid.NewString(),
		Type:     protocol.CredentialIssuanceResponseMessageType,
		ThreadID: basicMessage.ThreadID,
		Body:     protocol.IssuanceMessageBody{Credential: *cred},
		From:     basicMessage.To,
		To:       basicMessage.From,
	})
	return resp, err
}
