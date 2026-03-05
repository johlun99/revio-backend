package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	apimw "github.com/johlun99/revio/internal/api/middleware"
	"github.com/johlun99/revio/internal/repository"
)

type UsersHandler struct {
	queries *repository.Queries
}

func NewUsersHandler(pool *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{queries: repository.New(pool)}
}

func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListAdminUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	if rows == nil {
		rows = []repository.ListAdminUsersRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	role := repository.AdminRole(req.Role)
	if req.Role == "" {
		role = repository.AdminRoleModerator
	}
	if !role.Valid() {
		writeError(w, http.StatusBadRequest, "invalid role — must be superadmin, admin, or moderator")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.queries.CreateAdminUser(r.Context(), repository.CreateAdminUserParams{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	})
	if err != nil {
		writeError(w, http.StatusConflict, "email already in use")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := apimw.ClaimsFromContext(r.Context())

	var uid pgtype.UUID
	if err := uid.Scan(chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Prevent self-deletion
	if claims != nil && claims.UserID == uid.String() {
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	if err := h.queries.DeleteAdminUser(r.Context(), uid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
