package handlers

import (
	"encoding/json"
	"net/http"

	"ai-image-demo-backend/internal/models"
	"ai-image-demo-backend/internal/services"
	"ai-image-demo-backend/internal/utils"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Phone, req.Username, req.Password)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]models.User{"user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), req.Account, req.Password)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, result)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request, user models.User) {
	fullUser, err := h.auth.FindMe(r.Context(), user.ID)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "user not found")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]models.User{"user": fullUser})
}

type registerRequest struct {
	Phone    string `json:"phone"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}
