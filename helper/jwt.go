package helper

import (
	"time"
	"UAS_GO/app/models"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-min-32-characters-long")

func GenerateToken(user models.User) (string, error) {
	claims := models.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Validasi token JWT
func ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
