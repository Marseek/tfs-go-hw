package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	signedString = "lksjdkfjois9845784hug"
)

type myStr string

type messages struct {
	From    string
	Message string
	Time    time.Time
	To      string
}

var userNamesAndPass map[string]string
var toAllMessages []messages
var personalMessages map[string][]messages

func main() {
	personalMessages = make(map[string][]messages)
	userNamesAndPass = make(map[string]string)

	root := chi.NewRouter()
	root.Use(middleware.Logger)
	root.HandleFunc("/users/register", Register)

	r1 := chi.NewRouter()
	r1.Use(Auth)
	r1.Get("/", MessagesShow)            // Эндпойнт для отображения всех сообщений
	r1.Post("/", MessagesSend)           // Эндпойнт для отправки сообщений для всех пользователей
	r1.Get("/my", MyMessagesShow)        // Эндпойнт для запроса личных сообщений от авторизованного пользователя
	r1.Post("/personal", MyMessagesSend) // Эндпойнт для отправки личных сообщений от авторизованного пользователя
	root.Mount("/messages", r1)
	log.Fatal(http.ListenAndServe(":5000", root))
}

type Usr struct {
	Login  string
	Passwd string
}

type Message struct {
	Message string
	To      string
}

type TokenClaims struct {
	jwt.StandardClaims
	Login string `json:"Login"`
}

func Register(w http.ResponseWriter, r *http.Request) {
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
		return
	}
	userNamesAndPass[u.Login] = u.Passwd

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 1024).Unix(), IssuedAt: time.Now().Unix()}, u.Login})
	tokenString, _ := token.SignedString([]byte(signedString))
	log.Printf("%s user created and token generated\n", u.Login)
	w.Header().Add("token", tokenString)
}

func MessagesShow(w http.ResponseWriter, r *http.Request) {
	outString := getMessages(r.URL.Query())

	_, _ = w.Write([]byte(outString))
	w.WriteHeader(http.StatusAccepted)
}

func getMessages(a map[string][]string) string { // Вывод с пагинацией
	outString := ""
	paginate := true

	mapQuery := a
	fromArr, ok1 := mapQuery["from"]
	toArr, ok2 := mapQuery["to"]
	var from, to int
	var err1, err2 error
	if ok1 && ok2 {
		from, err1 = strconv.Atoi(fromArr[0])
		to, err2 = strconv.Atoi(toArr[0])
	}
	if !ok1 || !ok2 || err1 != nil || err2 != nil || from > to {
		paginate = false
	}
	if !paginate {
		for _, val := range toAllMessages {
			outString += fmt.Sprintf("%-14s: %-50s: %s\n", val.From, val.Message, val.Time.Format("15:04:05"))
		}
	} else {
		for i := from - 1; i < to && i < len(toAllMessages); i++ {
			outString += fmt.Sprintf("%-14s: %-50s: %s\n", toAllMessages[i].From, toAllMessages[i].Message, toAllMessages[i].Time.Format("15:04:05"))
		}
	}
	return outString
}

func MessagesSend(w http.ResponseWriter, r *http.Request) {
	userName, ok := r.Context().Value(myStr("ID")).(myStr)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var message Message
	err = json.Unmarshal(body, &message)
	if err != nil || message.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	toAllMessages = append(toAllMessages, messages{From: string(userName), Message: message.Message, Time: time.Now()})
	log.Printf("Message from user %s been sent to all\n", userName)
	w.WriteHeader(http.StatusAccepted)
}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")

		if tokenHeader == "" { // Токен отсутствует
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Printf("There is no Authorization Header. Error. Stop handling.")
			return
		}

		splitted := strings.Split(tokenHeader, " ") // Проверка соответствия формату `Bearer {token-body}`
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("There is invalid type of Authorization Header.")
			return
		}

		tokenPart := splitted[1]
		tk := &TokenClaims{}

		_, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(signedString), nil
		})
		if err != nil { // Неправильный или несуществующий токен
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Invalid token.")
			return
		}

		idCtx := context.WithValue(r.Context(), myStr("ID"), myStr(tk.Login))
		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}

func MyMessagesShow(w http.ResponseWriter, r *http.Request) {
	userName, ok := r.Context().Value(myStr("ID")).(myStr)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outString := fmt.Sprintf("%s's messages\n", userName)
	for _, val := range personalMessages[string(userName)] {
		outString += fmt.Sprintf("%-14s:%-50s: %s\n", val.From, val.Message, val.Time.Format("15:04:05"))
	}
	_, _ = w.Write([]byte(outString))

	w.WriteHeader(http.StatusAccepted)
}

func MyMessagesSend(w http.ResponseWriter, r *http.Request) {
	userName, ok := r.Context().Value(myStr("ID")).(myStr)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var message Message
	err = json.Unmarshal(body, &message)
	_, ok = userNamesAndPass[message.To]
	if err != nil || message.Message == "" || !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	personalMessages[message.To] = append(personalMessages[message.To], messages{From: string(userName), Message: message.Message, Time: time.Now()})
	log.Printf("%s user sent message to %s\n", userName, message.To)

	w.WriteHeader(http.StatusAccepted)
}
