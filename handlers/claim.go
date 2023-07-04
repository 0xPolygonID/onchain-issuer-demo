package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xPolygonID/onchain-issuer-demo/services"
	"github.com/go-chi/chi/v5"
	"github.com/iden3/iden3comm"
)

type Handlers struct {
	CredentialService services.OnChain
	Packager          *iden3comm.PackageManager
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

func (h *Handlers) IsRevokedClaim(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	nonce := chi.URLParam(r, "nonce")
	n, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	isRevoked, err := h.CredentialService.IsRevokedVC(r.Context(), issuer, uint64(n))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"mtp":{"existence": %t}}`, isRevoked)
}

func (h *Handlers) RevokeClaim(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	nonce := chi.URLParam(r, "nonce")
	n, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.CredentialService.RevokeVC(r.Context(), issuer, uint64(n)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetUserVCs(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	subject := r.URL.Query().Get("subject")
	schemaType := r.URL.Query().Get("schemaType")
	vcs, err := h.CredentialService.GetUsersVCs(r.Context(), issuer, subject, schemaType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vcs)
}

func (h *Handlers) GetUserVCByID(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	claimID := chi.URLParam(r, "claimId")
	vc, err := h.CredentialService.GetUserVCByID(r.Context(), issuer, claimID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vc)
}
