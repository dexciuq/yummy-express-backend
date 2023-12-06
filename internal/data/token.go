package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
	"reflect"
	"time"
)

var (
	ErrTokenExpired    = errors.New("token was expired")
	AccessTokenExpire  = time.Minute * 15
	RefreshTokenExpire = time.Hour * 24 * 30 //time.Minute * 5
)

type TokenData struct {
	UserID int64
	RoleID int64
}

type Token struct {
	ID           int64  `json:"-"`
	UserID       int64  `json:"-"`
	RoleID       int64  `json:"-"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token,omitempty"`
}

type TokenModel struct {
	DB *sql.DB
}

func GenerateTokens(userID int64, roleID int64) (*Token, error) {
	token := Token{
		UserID:       userID,
		RoleID:       roleID,
		RefreshToken: "",
		AccessToken:  "",
	}
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(RefreshTokenExpire).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_SK")))
	if err != nil {
		return nil, err
	}
	token.RefreshToken = refreshTokenString

	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(AccessTokenExpire).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("ACCESS_SK")))

	if err != nil {
		return nil, err
	}
	token.AccessToken = accessTokenString
	return &token, nil
}

func (t TokenModel) InsertToken(token *Token) error {
	query := `INSERT INTO tokens (refresh_token, user_id)
				 VALUES ($1, $2)
				 RETURNING id;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, &token.RefreshToken, &token.UserID).Scan(&token.ID)

	if err != nil {
		return err
	}

	return nil
}

func (t TokenModel) SaveToken(token *Token) error {
	tokenFromDb, err := t.FindTokenByUserId(token.UserID)
	if err != nil && !errors.Is(err, ErrRecordNotFound) {
		return err
	}

	if tokenFromDb != nil {
		t.UpdateToken(token)
		return nil
	}
	result := t.InsertToken(token)
	return result
}

func (t TokenModel) RemoveToken(refreshToken string) error {
	if refreshToken == "" {
		return ErrRecordNotFound
	}
	query := `DELETE FROM tokens WHERE refresh_token=$1;`

	result, err := t.DB.Exec(query, refreshToken)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (t TokenModel) UpdateToken(token *Token) error {
	query := `UPDATE tokens SET refresh_token = $1 WHERE user_id = $2`

	args := []any{token.RefreshToken, token.UserID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, args...)
	if err.Err() != nil {
		return err.Err()
	}

	return nil
}

func (t TokenModel) FindTokenByUserId(UserID int64) (*Token, error) {
	query := `SELECT id, user_id, refresh_token FROM tokens WHERE user_id = $1`

	var token Token
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, UserID).Scan(
		&token.ID,
		&token.UserID,
		&token.RefreshToken,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &token, nil
}

func (t TokenModel) FindToken(refreshToken string) (*Token, error) {
	query := `SELECT id, user_id, refresh_token FROM tokens WHERE refresh_token = $1`

	var token Token
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRowContext(ctx, query, refreshToken).Scan(
		&token.ID,
		&token.UserID,
		&token.RefreshToken,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &token, nil
}

func (t TokenModel) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := DecodeRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	fmt.Println(len(claims))
	if len(claims) == 0 {
		return "", ErrTokenExpired
	}
	exp := int64(claims["exp"].(float64))
	expUnix := time.Unix(exp, 0)
	fmt.Println(exp, reflect.TypeOf(exp))
	fmt.Println(expUnix, reflect.TypeOf(expUnix))
	fmt.Println(time.Now().After(expUnix), time.Now())
	if time.Now().After(expUnix) {
		fmt.Println("token expired")
		return "", ErrTokenExpired
	}
	userId := claims["user_id"].(float64)
	token, err := t.FindTokenByUserId(int64(userId))
	if err != nil {
		return "", err
	}

	if token.RefreshToken != refreshToken {
		return "", errors.New("wrong refresh token")
	}

	accessClaims := jwt.MapClaims{
		"user_id": claims["user_id"],
		"role_id": claims["role_id"],
		"exp":     time.Now().Add(AccessTokenExpire).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("ACCESS_SK")))

	if err != nil {
		return "", err
	}
	return accessTokenString, nil
}

func DecodeRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("REFRESH_SK")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token yoooo")
}

func DecodeAccessToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("incorrect signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("ACCESS_SK")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token yoooo")
}
