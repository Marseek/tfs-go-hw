package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/Marseek/tfs-go-hw/course/repository"
	"github.com/Marseek/tfs-go-hw/course/service"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestStop(t *testing.T) {
	t.Run("Set Stop", func(t *testing.T) {
		// Init Dependencies
		logger := log.New()
		rep := &repository.Repo{}
		serv := service.NewRobotService(rep, logger)
		serv.SetParams(100, 2, 1, "PI_XBTUSD", "buy")
		handler := NewParamsSetter(logger, serv)

		// Init Endpoint
		r := chi.NewRouter()
		r.Post("/api/stop", handler.Stop)

		// Create Request
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/stop",
			bytes.NewBufferString(""))

		// Make Request
		r.ServeHTTP(w, req)

		assert.Equal(t, 202, w.Code)
		assert.Equal(t, 0, serv.GetParams().Start)
		assert.Equal(t, "The signal to stop had been sent\n", w.Body.String())
	})
}

func TestSetAndStart(t *testing.T) {
	// Test Table
	type Test struct {
		Name         string
		InBody       string
		ExpectStCode int
		ExpectBody   string
	}
	tests := [...]Test{
		{"Status Accepted", `{"start":1, "ticker":"PI_XBTUSD", "size":2, "profit":0.05, "side":"buy"}`, 200, "Parameters had been set\n"},
		{"Unmarshall error", `{"side":2}`, 400, "Json unmarshall error"},
		{"Start param error", `{"start":-1, "ticker":"PI_XBTUSD", "size":2, "profit":0.05, "side":"buy"}`, 400, "Bad params: 'start' option must be '1' or '0'"},
	}

	// Init Dependencies
	logger := log.New()
	rep := &repository.Repo{}
	serv := service.NewRobotService(rep, logger)
	handler := NewParamsSetter(logger, serv)

	// Init Endpoint
	r := chi.NewRouter()
	r.Post("/api/", handler.SetAndStart)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/",
				bytes.NewBufferString(test.InBody))

			// Make Request
			r.ServeHTTP(w, req)

			assert.Equal(t, test.ExpectStCode, w.Code)
			assert.Equal(t, test.ExpectBody, w.Body.String())
		})
	}
}

func TestSetParam(t *testing.T) {
	// Test Table
	type Test struct {
		Name         string
		InBody       string
		ExpectStCode int
		ExpectBody   string
	}
	tests := [...]Test{
		{"Status Accepted", `{"start":1, "ticker":"PI_XBTUSD", "size":2, "profit":0.05, "side":"buy"}`, 200, "Parameters had been set\n"},
		{"Unmarshall error", `{"side":2}`, 400, "Json unmarshall error\n"}, //
		{"Size error", `{"start":1, "ticker":"PI_XBTUSD", "size":-2, "profit":0.05, "side":"buy"}`, 400, "Bad params: 'size' option must be more than 0"},
		{"Profit param error", `{"start":1, "ticker":"PI_XBTUSD", "size":2, "profit":-0.05, "side":"buy"}`, 400, "Bad params: 'profit' must be more than 0"},
	}
	// Init Dependencies
	logger := log.New()
	rep := &repository.Repo{}
	serv := service.NewRobotService(rep, logger)
	handler := NewParamsSetter(logger, serv)

	// Init Endpoint
	r := chi.NewRouter()
	r.Post("/api/set", handler.SetParam)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/set",
				bytes.NewBufferString(test.InBody))

			// Make Request
			r.ServeHTTP(w, req)

			assert.Equal(t, test.ExpectStCode, w.Code)
			assert.Equal(t, test.ExpectBody, w.Body.String())
		})
	}
}

func TestStart(t *testing.T) {
	// Test Table
	type Test struct {
		Name          string
		In            domain.Options
		ExpectStCode  int
		ExpectParamSt int
		ExpectBody    string
	}
	tests := [...]Test{
		{Name: "Status Accepted", In: domain.Options{Start: 10, Ticker: "PI_XBTUSD", Size: 1, Profit: 0.1, Side: "buy"}, ExpectStCode: 202, ExpectParamSt: 1, ExpectBody: "The signal to start had been sent\n"},
		{Name: "Bad Params", In: domain.Options{Start: 10, Ticker: "Wrong_Ticker", Size: 1, Profit: 0.1, Side: "buy"}, ExpectStCode: 400, ExpectParamSt: 10, ExpectBody: "Bad params: 'ticker' option must be 'PI_XBTUSD' or 'PI_ETHUSD' or 'PI_LTCUSD' or 'PI_XRPUSD' or 'PI_BCHUSD'"},
	}
	// Init Dependencies
	logger := log.New()
	rep := &repository.Repo{}
	serv := service.NewRobotService(rep, logger)
	handler := NewParamsSetter(logger, serv)

	// Init Endpoint
	r := chi.NewRouter()
	r.Post("/api/start", handler.Start)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Create Request
			serv.SetParams(test.In.Start, test.In.Size, test.In.Profit, test.In.Ticker, test.In.Side)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/start",
				bytes.NewBufferString(""))

			// Make Request
			r.ServeHTTP(w, req)

			assert.Equal(t, test.ExpectStCode, w.Code)
			assert.Equal(t, test.ExpectParamSt, serv.GetParams().Start)
			assert.Equal(t, test.ExpectBody, w.Body.String())
		})
	}
}
