package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	RoleID   int    `json:"role_id"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(adminID int64, username, realName string, roleID int) (string, error) {
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		RealName: realName,
		RoleID:   roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret()))
}

func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(getJWTSecret()), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func getJWTSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}
	return "steellink-admin-dev-secret"
}
