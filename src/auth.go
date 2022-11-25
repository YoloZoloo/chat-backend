package src

import (
	model "chat-backend/model"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/astaxie/session"
	_ "github.com/go-sql-driver/mysql"
)

type JWTClaim struct {
	UniqueID uint   `json:"id"`
	UserName string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateJWT(user model.User) (tokenString string, err error) {
	var jwtKey = []byte(os.Getenv("GO_CHAT_JWT_KEY"))
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		UniqueID: uint(user.ID),
		UserName: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func ValidateToken(signedToken string) (uint, error) {
	var jwtKey = []byte(os.Getenv("GO_CHAT_JWT_KEY"))
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*JWTClaim)
	id := claims.UniqueID
	if !ok {
		err = errors.New("couldn't parse claims")
		return 0, err
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return 0, err
	}
	return id, err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(credentials UserCredentials) (model.User, error) {
	db, err := model.DbInit()

	if err != nil {
		return model.User{}, err
	}
	var user model.User
	result := db.Take(&user, "username = ?", credentials.UserName)
	if result.RowsAffected == 0 || result.Error != nil {
		return model.User{}, errors.New("no such user")
	}
	fmt.Println(result)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		return model.User{}, errors.New("invalid username or password")
	}

	return user, nil
}

