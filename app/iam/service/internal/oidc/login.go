package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/pkg/helpers"
	"github.com/Servora-Kit/servora/pkg/logger"
	"github.com/Servora-Kit/servora/pkg/redis"
)

const loginTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Sign In — Servora</title>
  <style>
    *{box-sizing:border-box;margin:0;padding:0}
    body{font-family:system-ui,-apple-system,sans-serif;background:#f5f5f5;display:flex;align-items:center;justify-content:center;min-height:100vh}
    .card{background:#fff;border-radius:8px;box-shadow:0 2px 8px rgba(0,0,0,.08);padding:2rem;width:100%;max-width:380px}
    h1{font-size:1.25rem;margin-bottom:1.5rem;text-align:center;color:#111}
    .field{margin-bottom:1rem}
    label{display:block;font-size:.875rem;color:#555;margin-bottom:.25rem}
    input{width:100%;padding:.5rem .75rem;border:1px solid #ddd;border-radius:4px;font-size:.875rem}
    input:focus{outline:none;border-color:#4f46e5}
    .btn{width:100%;padding:.625rem;background:#4f46e5;color:#fff;border:none;border-radius:4px;font-size:.875rem;cursor:pointer;margin-top:.5rem}
    .btn:hover{background:#4338ca}
    .err{color:#dc2626;font-size:.8rem;margin-bottom:1rem;text-align:center}
  </style>
</head>
<body>
  <div class="card">
    <h1>Sign In</h1>
    {{if .Error}}<p class="err">{{.Error}}</p>{{end}}
    <form method="POST" action="/login">
      <input type="hidden" name="authRequestID" value="{{.AuthRequestID}}">
      <div class="field">
        <label for="email">Email</label>
        <input id="email" name="email" type="email" required autofocus>
      </div>
      <div class="field">
        <label for="password">Password</label>
        <input id="password" name="password" type="password" required>
      </div>
      <button class="btn" type="submit">Sign In</button>
    </form>
  </div>
</body>
</html>`

var loginTemplate = template.Must(template.New("login").Parse(loginTmpl))

// LoginHandler handles the SSR login page (GET/POST /login).
type LoginHandler struct {
	authnRepo biz.AuthnRepo
	redis     *redis.Client
	log       *logger.Helper
}

// NewLoginHandler builds a handler that authenticates users and marks OIDC auth requests done.
func NewLoginHandler(authnRepo biz.AuthnRepo, rdb *redis.Client, l logger.Logger) *LoginHandler {
	return &LoginHandler{
		authnRepo: authnRepo,
		redis:     rdb,
		log:       logger.NewHelper(l, logger.WithModule("oidc/login/iam-service")),
	}
}

// ServeHTTP dispatches to GET (render form) or POST (handle form submission).
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.renderLogin(w, r)
	case http.MethodPost:
		h.handleLogin(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *LoginHandler) renderLogin(w http.ResponseWriter, r *http.Request) {
	authRequestID := r.URL.Query().Get("authRequestID")
	if authRequestID == "" {
		http.Error(w, "missing authRequestID", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = loginTemplate.Execute(w, map[string]string{
		"AuthRequestID": authRequestID,
		"Error":         "",
	})
}

func (h *LoginHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	authRequestID := r.FormValue("authRequestID")
	email := r.FormValue("email")
	password := r.FormValue("password")

	callbackURL, err := h.authenticate(r.Context(), authRequestID, email, password)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = loginTemplate.Execute(w, map[string]string{
			"AuthRequestID": authRequestID,
			"Error":         err.Error(),
		})
		return
	}
	http.Redirect(w, r, callbackURL, http.StatusFound)
}

// LoginCompleteHandler serves the JSON API at POST /login/complete.
type LoginCompleteHandler struct {
	lh *LoginHandler
}

// NewLoginCompleteHandler builds the API handler that returns callbackURL in JSON.
func NewLoginCompleteHandler(lh *LoginHandler) *LoginCompleteHandler {
	return &LoginCompleteHandler{lh: lh}
}

func (h *LoginCompleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AuthRequestID string `json:"authRequestID"`
		Email         string `json:"email"`
		Password      string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	callbackURL, err := h.lh.authenticate(r.Context(), req.AuthRequestID, req.Email, req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"callbackURL": callbackURL})
}

func (h *LoginHandler) authenticate(ctx context.Context, authRequestID, email, password string) (string, error) {
	if authRequestID == "" || email == "" || password == "" {
		return "", fmt.Errorf("missing required fields")
	}

	user, err := h.authnRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", fmt.Errorf("invalid email or password")
		}
		h.log.Errorf("get user by email: %v", err)
		return "", fmt.Errorf("internal error")
	}

	if !helpers.BcryptCheck(password, user.Password) {
		return "", fmt.Errorf("invalid email or password")
	}

	key := "oidc:auth_request:" + authRequestID
	data, err := h.redis.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("auth request not found or expired")
	}

	var reqData map[string]any
	if err := json.Unmarshal([]byte(data), &reqData); err != nil {
		return "", fmt.Errorf("internal error")
	}
	reqData["user_id"] = user.ID
	reqData["auth_time"] = time.Now().UTC().Format(time.RFC3339Nano)
	reqData["done"] = true

	updated, _ := json.Marshal(reqData)
	if err := h.redis.Set(ctx, key, string(updated), 10*time.Minute); err != nil {
		return "", fmt.Errorf("internal error")
	}

	return fmt.Sprintf("/authorize/callback?id=%s", authRequestID), nil
}
