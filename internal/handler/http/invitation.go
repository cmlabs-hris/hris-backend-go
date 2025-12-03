package http

import (
	"encoding/json"
	"net/http"

	"github.com/cmlabs-hris/hris-backend-go/internal/domain/invitation"
	"github.com/cmlabs-hris/hris-backend-go/internal/handler/http/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type InvitationHandler interface {
	// Public endpoint - view invitation details
	GetInvitationByToken(w http.ResponseWriter, r *http.Request)
	// Authenticated endpoints
	ListMyInvitations(w http.ResponseWriter, r *http.Request)
	AcceptInvitation(w http.ResponseWriter, r *http.Request)
}

type invitationHandlerImpl struct {
	invitationService invitation.InvitationService
}

func NewInvitationHandler(invitationService invitation.InvitationService) InvitationHandler {
	return &invitationHandlerImpl{
		invitationService: invitationService,
	}
}

// GetInvitationByToken implements InvitationHandler - public endpoint
func (h *invitationHandlerImpl) GetInvitationByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		response.BadRequest(w, "Token is required", nil)
		return
	}

	result, err := h.invitationService.GetByToken(r.Context(), token)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, result)
}

// ListMyInvitations implements InvitationHandler - lists pending invitations for authenticated user
func (h *invitationHandlerImpl) ListMyInvitations(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		response.Unauthorized(w, "Email not found in token")
		return
	}

	results, err := h.invitationService.ListMyInvitations(r.Context(), email)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.Success(w, results)
}

// AcceptInvitation implements InvitationHandler - accept an invitation
func (h *invitationHandlerImpl) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	_, claims, _ := jwtauth.FromContext(r.Context())

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		response.Unauthorized(w, "User ID not found in token")
		return
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		response.Unauthorized(w, "Email not found in token")
		return
	}

	var req invitation.AcceptRequest
	req.Token = token
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request format", nil)
		return
	}

	if err := req.Validate(); err != nil {
		response.HandleError(w, err)
		return
	}

	result, err := h.invitationService.Accept(r.Context(), req.Token, userID, email)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.SuccessWithMessage(w, "Invitation accepted successfully", result)
}
