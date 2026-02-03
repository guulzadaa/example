package handlers

import (
	"encoding/json"
	"net/http"

	"bookstore/internal/logic"
)

type AuthHandler struct {
	service *logic.AuthService
}

func NewAuthHandler(service *logic.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{"error": "method not allowed"},
		)
		return
	}

	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": "invalid JSON"},
		)
		return
	}

	if err := h.service.Register(in.Email, in.Password); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": err.Error()},
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]string{"message": "registered"},
	)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(
			w,
			http.StatusMethodNotAllowed,
			map[string]string{"error": "method not allowed"},
		)
		return
	}

	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{"error": "invalid JSON"},
		)
		return
	}

	token, err := h.service.Login(in.Email, in.Password)
	if err != nil {
		writeJSON(
			w,
			http.StatusUnauthorized,
			map[string]string{"error": err.Error()},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]string{"token": token},
	)
}
