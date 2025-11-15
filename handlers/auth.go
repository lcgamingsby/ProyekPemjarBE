package handlers

import (
	"database/sql"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"backend/middleware"
	"backend/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	DB        *sql.DB
	JWTSecret string
	ExpireH   int
}

type registerPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(db *sql.DB, jwtSecret string, expireH int) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret, ExpireH: expireH}
}

// Register: POST /api/register
func (h *AuthHandler) Register(c *gin.Context) {
	var body registerPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check existing
	var exists int
	err := h.DB.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?", body.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	res, err := h.DB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", body.Email, string(hash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert error"})
		return
	}
	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "email": body.Email})
}

// Login: POST /api/login
func (h *AuthHandler) Login(c *gin.Context) {

	var body loginPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.DB.QueryRow("SELECT id, email, password_hash, created_at FROM users WHERE email = ?", body.Email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// create JWT
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     now.Add(time.Hour * time.Duration(h.ExpireH)).Unix(),
		"iat":     now.Unix(),
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed})
}

// Protected test route example
func (h *AuthHandler) Sessions(c *gin.Context) {
	// middleware put claims in context
	claims, _ := c.Get(middleware.JwtClaimsKey)
	// dummy response
	c.JSON(http.StatusOK, gin.H{
		"sessions": []gin.H{
			{"id": "sess-1", "name": "Example Session", "status": "active"},
		},
		"claims": claims,
	})
}
