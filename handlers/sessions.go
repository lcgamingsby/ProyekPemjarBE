package handlers

import (
	"database/sql"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"backend/middleware"
	"backend/models"
)

// SessionsHandler holds DB connection
type SessionsHandler struct {
	DB *sql.DB
}

func NewSessionsHandler(db *sql.DB) *SessionsHandler {
	return &SessionsHandler{DB: db}
}

// helper: generate short code (6 chars alnum)
func generateCode(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(letters[r.Intn(len(letters))])
	}
	return b.String()
}

// POST /api/sessions  (protected)
type createSessionPayload struct {
	Name             string `json:"name" binding:"required"`
	MaxCollaborators int    `json:"maxCollaborators" binding:"omitempty,min=1"`
}

func (h *SessionsHandler) CreateSession(c *gin.Context) {
	var payload createSessionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.MaxCollaborators == 0 {
		payload.MaxCollaborators = 5
	}

	// get owner id from JWT claims if present
	claims := c.MustGet(middleware.JwtClaimsKey).(jwt.MapClaims)
	userID := int64(claims["user_id"].(float64))

	// ensure unique code (retry few times)
	var code string
	for i := 0; i < 5; i++ {
		code = generateCode(6)
		var exists int
		err := h.DB.QueryRow("SELECT COUNT(1) FROM sessions WHERE code = ?", code).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if exists == 0 {
			break
		}
	}
	if code == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed generate code"})
		return
	}

	// insert
	res, err := h.DB.Exec(
		"INSERT INTO sessions (code, name, owner_id, max_collaborators, status) VALUES (?, ?, ?, ?, ?)",
		code, payload.Name, userID, payload.MaxCollaborators, "active",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert error"})
		return
	}
	id, _ := res.LastInsertId()

	c.JSON(http.StatusCreated, gin.H{
		"id":                id,
		"code":              code,
		"name":              payload.Name,
		"maxCollaborators":  payload.MaxCollaborators,
	})
}

// GET /api/sessions (protected) — list sessions
func (h *SessionsHandler) ListSessions(c *gin.Context) {
	claims := c.MustGet(middleware.JwtClaimsKey).(jwt.MapClaims)
	userID := int64(claims["user_id"].(float64))
	//fmt.Println("USER ID DARI JWT:", userID)

	rows, err := h.DB.Query(`
		SELECT id, code, name, owner_id, max_collaborators, status, created_at, updated_at
		FROM sessions
		WHERE owner_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		var owner sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &owner, &s.MaxCollaborators, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db scan error"})
			return
		}
		if owner.Valid {
			s.OwnerID = uint64(owner.Int64)
		}
		sessions = append(sessions, s)
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// GET /api/sessions/join/:code (public) — find session by code
func (h *SessionsHandler) GetSessionByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}
	var s models.Session
	var owner sql.NullInt64
	err := h.DB.QueryRow("SELECT id, code, name, owner_id, max_collaborators, status, created_at, updated_at FROM sessions WHERE code = ?", code).
		Scan(&s.ID, &s.Code, &s.Name, &owner, &s.MaxCollaborators, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if owner.Valid {
		s.OwnerID = uint64(owner.Int64)
	}
	// You could optionally check status or collaborator count here
	c.JSON(http.StatusOK, gin.H{"session": s})
}
