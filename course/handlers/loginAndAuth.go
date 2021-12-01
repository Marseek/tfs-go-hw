package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	signedString = "lksjdkfjois9845784hug"
)

type Usr struct {
	Login  string `json:"login"`
	Passwd string `json:"passwd"`
}

type TokenClaims struct {
	jwt.StandardClaims
	Login string `json:"Login"`
}

func (p *SetParams) Login(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var u Usr
	err = json.Unmarshal(d, &u)
	if err != nil || u.Login == "" || u.Passwd == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Can't unmarshall data or empty username or password")
		return
	}

	usersMap := p.Service.GetUsersMap("users.json")
	if usersMap[u.Login] != u.Passwd {
		_, _ = io.WriteString(w, "Incorrect username or password")
		return
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&TokenClaims{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 1024).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			u.Login,
		},
	)
	tokenString, _ := token.SignedString([]byte(signedString))
	w.Header().Add("token", tokenString)
	_, _ = io.WriteString(w, "Token generated\n")
}

func (p *SetParams) Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" { // Токен отсутствует
			w.WriteHeader(http.StatusUnauthorized)
			p.logger.Debugf("There is no Authorization Header. Error. Stop handling.")
			return
		}

		splitted := strings.Split(tokenHeader, " ") // Проверка соответствия формату `Bearer {token-body}`
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			p.logger.Debugf("There is invalid type of Authorization Header.")
			return
		}

		tokenPart := splitted[1]
		tk := &TokenClaims{}

		_, err := jwt.ParseWithClaims(tokenPart, tk, func(*jwt.Token) (interface{}, error) {
			return []byte(signedString), nil
		})
		if err != nil { // Неправильный или несуществующий токен
			w.WriteHeader(http.StatusBadRequest)
			p.logger.Debugf("Invalid token.")
			return
		}

		handler.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
