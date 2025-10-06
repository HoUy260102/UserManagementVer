package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService struct {
	SecretKey string
	Issuer    string
}

func NewJwtService(secretKey string, issuer string) *JwtService {
	return &JwtService{SecretKey: secretKey, Issuer: issuer}
}

type JwtCustomClaim struct {
	Email string
	Type  string
	Role  string
	jwt.RegisteredClaims
}

func (j *JwtService) GenerateJwt(email string, duration int, typeToken string) (string, *JwtCustomClaim, error) {
	tokenId, _ := uuid.NewRandom()
	claims := &JwtCustomClaim{
		Email: email,
		Type:  typeToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenId.String(),
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(duration) * time.Second)),
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", nil, err
	}
	return tok, claims, nil
}

func (j *JwtService) ExtractCustomClaims(tokenStr string) (*JwtCustomClaim, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtCustomClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtCustomClaim); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(t_ *jwt.Token) (interface{}, error) {
		if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", t_.Header["alg"])
		}
		return []byte(j.SecretKey), nil
	})

	return token, err
}
