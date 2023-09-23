package util

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/anazibinurasheed/dmart-auth-svc/internal/config"
	"github.com/anazibinurasheed/dmart-auth-svc/internal/pb"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Returns password hash
func HashPassword(password string) (string, error) {
	byte, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if HasError(context.Background(), nil) {
		return "", err
	}

	return string(byte), nil
}

func CompareHashAndPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

func GetLoginMethod(req *pb.UserLoginRequest) (method string) {
	len := len(req.LoginInput)
	Logger("reached")
	if len > 10 && string(req.LoginInput[(len)-10:]) == "@gmail.com" {
		method = "email"
		Logger(string(req.LoginInput[(len)-11:]))
		return
	}

	if _, err := strconv.Atoi(req.LoginInput); err != nil {
		Logger("username")
		method = "username"
		return
	}

	method = "phone"

	return
}

type jwtClaims struct {
	ExpiresAt int64
	Issuer    string
	UserID    string
	jwt.Claims
}

// GenerateToken can generate the token for user and admin with different secret key, should need to specify the role
func GenerateToken(userID string, role string) (signedToken string, err error) {
	claims := &jwtClaims{
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).Unix(),
		Issuer:    "auth-svc",
		UserID:    userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	switch role {
	case "admin":
		signedToken, err = token.SignedString([]byte(config.GetConfig().AdminJwtSecret))
	case "user":
		signedToken, err = token.SignedString([]byte(config.GetConfig().UserJwtSecret))
	default:
		err = errors.New("invalid role argument")
	}

	if HasError(context.Background(), err) {
		return "", err
	}

	return signedToken, nil

}

func ValidateToken(signedToken string, role string) (claims *jwtClaims, err error) {
	var secretKey string

	switch role {
	case "admin":
		secretKey = config.GetConfig().AdminJwtSecret
	case "user":
		secretKey = config.GetConfig().UserJwtSecret
	default:
		err = errors.New("invalid role argument")
	}
	if HasError(context.Background(), err) {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(signedToken, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if HasError(context.Background(), err) {
		return nil, err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}

	return
}
