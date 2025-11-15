package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const JwtClaimsKey = "jwt_claims"

func JWTAuth(secret string) gin.HandlerFunc {
	//fmt.Println("VERIFY USE SECRET =", secret)
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		var tokenStr string
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		} else {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		//fmt.Println("🔥 AUTH HEADER DITERIMA BACKEND:", c.Request.Header)
		//fmt.Println("🔥 AUTH RAW:", c.GetHeader("Authorization"))


		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		//fmt.Println("TOKEN DITERIMA:", tokenStr)
		c.Set(JwtClaimsKey, token.Claims)
		c.Next()
	}
}
