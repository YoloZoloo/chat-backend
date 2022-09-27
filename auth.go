package main

import (
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/astaxie/session"
	_ "github.com/go-sql-driver/mysql"
)

type JWTClaim struct {
	UniqueID int16  `json:"id"`
	UserID   string `json:"user_id"`
	jwt.StandardClaims
}

func GenerateJWT(user_id string, id int16) (tokenString string, err error) {
	var jwtKey = []byte(os.Getenv("GO_CHAT_JWT_KEY"))
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		UniqueID: id,
		UserID:   user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func ValidateToken(signedToken string) (int16, error) {
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
	var id int16
	id = claims.UniqueID
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

func CheckPassword(user_id string, password string) (int16, error) {
	db, err := OpenDB()

	if err != nil {
		return 0, err
	}

	queryString := "select id, user_id, password from user_m where user_id = ?"

	stmt, err := db.Prepare(queryString)

	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	userId := ""
	accountPassword := ""
	var id int16

	err = stmt.QueryRow(user_id).Scan(&id, &userId, &accountPassword)

	if err != nil {

		if err == sql.ErrNoRows {
			return 0, errors.New("Invalid username or password.\r\n")
		}

		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(accountPassword), []byte(password))

	if err != nil {
		return 0, errors.New("Invalid username or password.\r\n")
	}
	return id, nil
}
