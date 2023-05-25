package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iden3/issuer-on-chain-backend/services"
)

type Handlers struct {
	CredentialService services.OnChain
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
