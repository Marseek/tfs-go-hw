package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Marseek/tfs-go-hw/course/repository"
	"github.com/Marseek/tfs-go-hw/course/service"
	mock_service "github.com/Marseek/tfs-go-hw/course/service/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	// Test Table
	type mockBehavior func(r *mock_service.MockrepoInterface)
	type Test struct {
		Name         string
		InBody       string
		mockBehavior mockBehavior
		ExpectCode   int
		ExpectBody   string
	}
	tests := [...]Test{
		{
			"Right format",
			`{"login": "jlexie", "passwd": "passwd"}`,
			func(r *mock_service.MockrepoInterface) {
				r.EXPECT().GetUsersMap("users.json").Return(map[string]string{"jlexie": "passwd"})
			},
			200,
			"Token generated\n",
		},
		{
			"Invalid format",
			`{"name": "jlexie", "age": "passwd"}`,
			func(r *mock_service.MockrepoInterface) {},
			400,
			"Can't unmarshall data or empty username or password",
		},
		{
			"Wrong passwd",
			`{"login": "jlexie", "passswd": "passwd"}`,
			func(r *mock_service.MockrepoInterface) {},
			400,
			"Can't unmarshall data or empty username or password",
		},
		{
			"Empty login and pass",
			`{"login": "", "passswd": ""}`,
			func(r *mock_service.MockrepoInterface) {},
			400,
			"Can't unmarshall data or empty username or password",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			logger := log.New()
			repo := mock_service.NewMockrepoInterface(c)
			test.mockBehavior(repo)

			serv := service.NewRobotService(repo, logger)
			handler := NewParamsSetter(logger, serv)

			// Init Endpoint
			r := chi.NewRouter()
			r.Post("/sign-up", handler.Login)

			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/sign-up",
				bytes.NewBufferString(test.InBody))

			// Make Request
			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.ExpectCode)
			assert.Equal(t, w.Body.String(), test.ExpectBody)
		})
	}
}

func TestAuth(t *testing.T) {
	// Test Table
	type Test struct {
		Name   string
		In     string
		Expect int
	}
	okToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDEzNjMzNzcsImlhdCI6MTYzNzY3Njk3NywiTG9naW4iOiJqbGV4aWUifQ.JUr3hVS4c0-HrbzKCMCJrLbAn34TVg3NKXXRdXU-e2g"
	tests := [...]Test{
		{"Invalid Token Format", "arihgeir", 400},
		{"Invalid Token", "Bearer lskfjl", 400},
		{"Status Accepted", okToken, 200},
	}
	// Init Dependencies
	logger := log.New()
	rep := &repository.Repo{}
	serv := service.NewRobotService(rep, logger)
	handler := NewParamsSetter(logger, serv)

	// Init Endpoint
	r := chi.NewRouter()
	r.Use(handler.Auth)
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello, World!"))
	})

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/login",
				bytes.NewBufferString("Simple text"))
			req.Header.Add("Authorization", test.In)

			// Make Request
			r.ServeHTTP(w, req)

			assert.Equal(t, w.Code, test.Expect)
		})
	}
}
