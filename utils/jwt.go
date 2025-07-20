package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/okanay/backend-template/configs"
	"github.com/okanay/backend-template/types"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	types.TokenClaims
}

// GenerateAccessToken, verilen minimal 'claims' ile yeni bir Access Token oluşturur.
func GenerateAccessToken(claims types.TokenClaims) (string, error) {
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_ACCESS_SECRET environment variable not set")
	}

	// Token'ın geçerlilik süresini ve diğer standart talepleri ayarla.
	// 'Subject' olarak artık 'Email' yerine 'ID' kullanmak daha doğru ve güvenlidir.
	tokenClaims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(configs.ACCESS_TOKEN_DURATION)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    configs.JWT_ISSUER,
			Subject:   claims.ID.String(), // Konu olarak User ID kullanılıyor.
		},
		TokenClaims: claims, // types.TokenClaims { ID, Role }
	}

	// Token'ı imzala.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func ValidateAccessToken(tokenString string) (*types.TokenClaims, error) {
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_ACCESS_SECRET environment variable not set")
	}

	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	// Hata kontrolü ve token geçerliliği kontrolü.
	if err != nil {
		return nil, err // Hata, token'ın süresinin dolduğunu da içerebilir (jwt.ErrTokenExpired).
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return &claims.TokenClaims, nil
}

func IsTokenExpired(tokenString string) (bool, error) {
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return true, errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return true, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return true, errors.New("invalid token")
	}

	return false, nil
}

func ExtractClaims(tokenString string) (*types.TokenClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return &claims.TokenClaims, nil
	}

	return nil, errors.New("invalid claims format")
}

func GenerateRefreshToken() (string, error) {
	token, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

func ShouldRefreshToken(tokenString string) (bool, error) {
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	if secretKey == "" {
		return false, errors.New("JWT_ACCESS_SECRET environment variable is not set")
	}

	claims := &JWTClaims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	expiresAt := claims.ExpiresAt.Time
	issuedAt := claims.IssuedAt.Time
	totalDuration := expiresAt.Sub(issuedAt)
	remainingDuration := time.Until(expiresAt)

	return remainingDuration < (totalDuration / 4), nil
}
