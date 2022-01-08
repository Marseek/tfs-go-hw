package handlers

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

	"github.com/Marseek/tfs-go-hw/home_work_04/domain"
	"github.com/Marseek/tfs-go-hw/home_work_04/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
)

type Chat struct {
	Rep repository.Repository
}

func NewChatHandlers(rep repository.Repository) *Chat {
	return &Chat{
		Rep: rep,
	}
}

func (p *Chat) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/users/register", p.Register)

	r1 := chi.NewRouter()
	r1.Use(p.Auth)
	r1.Get("/", p.MessagesShow)            // Эндпойнт для отображения всех сообщений
	r1.Post("/", p.MessagesSend)           // Эндпойнт для отправки сообщений для всех пользователей
	r1.Get("/my", p.MyMessagesShow)        // Эндпойнт для запроса личных сообщений от авторизованного пользователя
	r1.Post("/personal", p.MyMessagesSend) // Эндпойнт для отправки личных сообщений от авторизованного пользователя
	r.Mount("/messages", r1)
	return r
}

func (p *Chat) MessagesShow(w http.ResponseWriter, r *http.Request) { // Вывод с пагинацией
	outString := ""
	paginate := true

	mapQuery := r.URL.Query()
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
		for _, val := range p.Rep.ShowAll() {
			outString += fmt.Sprintf("%-14s: %-50s: %s\n", val.From, val.Message, val.Time.Format("15:04:05"))
		}
	} else {
		messages := p.Rep.ShowAll()
		for i := from - 1; i < to && i < len(messages); i++ {
			outString += fmt.Sprintf("%-14s: %-50s: %s\n", messages[i].From, messages[i].Message, messages[i].Time.Format("15:04:05"))
		}
	}

	_, _ = w.Write([]byte(outString))
}

func (p *Chat) MessagesSend(w http.ResponseWriter, r *http.Request) {
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
	p.Rep.SendToAll(domain.Messages{From: string(userName), Message: message.Message, Time: time.Now()})

	log.Printf("Message from user %s been sent to all\n", userName)
	w.WriteHeader(http.StatusAccepted)
}

func (p *Chat) MyMessagesShow(w http.ResponseWriter, r *http.Request) {
	userName, ok := r.Context().Value(myStr("ID")).(myStr)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outString := p.Rep.ShowMy(string(userName))
	_, _ = w.Write([]byte(outString))
}

func (p *Chat) MyMessagesSend(w http.ResponseWriter, r *http.Request) {
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
	ok = p.Rep.UserIsPresent(message.To)
	if err != nil || message.Message == "" || !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.Rep.SendToSmb(domain.Messages{From: string(userName), Message: message.Message, Time: time.Now()}, message.To)
	log.Printf("%s user sent message to %s\n", userName, message.To)

	w.WriteHeader(http.StatusAccepted)
}

func (p *Chat) Auth(handler http.Handler) http.Handler {
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

type myStr string

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

func (p *Chat) Register(w http.ResponseWriter, r *http.Request) {
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

	p.Rep.RegisterUser(u.Login, u.Passwd)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Hour * 1024).Unix(), IssuedAt: time.Now().Unix()}, u.Login})
	tokenString, _ := token.SignedString([]byte(signedString))
	log.Printf("%s user created and token generated\n", u.Login)
	w.Header().Add("token", tokenString)
}

const (
	signedString = "lksjdkfjois9845784hug"
)
