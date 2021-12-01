package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Marseek/tfs-go-hw/course/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type RobotService interface {
	SetParams(start, size int, profit float32, ticker, side string)
	SetParamsWithoutStart(size int, profit float32, ticker, side string)
	GetParams() domain.Options
	SetStart(start int)
	GetUsersMap(string) map[string]string
}

type SetParams struct {
	Service RobotService
	logger  logrus.FieldLogger
}

func NewParamsSetter(logger logrus.FieldLogger, service RobotService) *SetParams {
	return &SetParams{
		Service: service,
		logger:  logger,
	}
}

func (p *SetParams) Routes() chi.Router {
	root := chi.NewRouter()
	root.Use(middleware.Logger)

	root.HandleFunc("/login", p.Login)

	r := chi.NewRouter()
	// r.Use(p.Auth)
	r.Post("/", p.SetAndStart)
	r.Post("/set", p.SetParam)
	r.Post("/start", p.Start)
	r.Post("/stop", p.Stop)
	root.Mount("/api", r)

	return root
}

func (p *SetParams) Start(w http.ResponseWriter, r *http.Request) {
	par := p.Service.GetParams()
	err := checkInput(par)
	if err != nil {
		p.logger.WithError(err).Error("Error, while setting parm's")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad params: "+err.Error())
		return
	}
	p.Service.SetStart(1)
	w.WriteHeader(http.StatusAccepted)
	_, _ = io.WriteString(w, "The signal to start had been sent\n")
}

func (p *SetParams) Stop(w http.ResponseWriter, r *http.Request) {
	p.Service.SetStart(0)
	w.WriteHeader(http.StatusAccepted)
	_, _ = io.WriteString(w, "The signal to stop had been sent\n")
}

func (p *SetParams) SetAndStart(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var options domain.Options
	err = json.Unmarshal(body, &options)
	if err != nil {
		p.logger.Println("Unmarshall error")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Json unmarshall error")
		return
	}

	err = checkInputWithStart(options)
	if err != nil {
		p.logger.WithError(err).Error("Error, while setting parms")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad params: "+err.Error())
		return
	}
	_, _ = io.WriteString(w, "Parameters had been set\n")
	p.Service.SetParams(options.Start, options.Size, options.Profit, options.Ticker, options.Side)
}

func (p *SetParams) SetParam(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var options domain.Options
	err = json.Unmarshal(body, &options)
	if err != nil {
		p.logger.Println("Unmarshall error")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Json unmarshall error\n")
		return
	}
	err = checkInput(options)
	if err != nil {
		p.logger.WithError(err).Error("Error, while setting parm's")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Bad params: "+err.Error())
		return
	}
	_, _ = io.WriteString(w, "Parameters had been set\n")
	p.Service.SetParamsWithoutStart(options.Size, options.Profit, options.Ticker, options.Side)
}

func (p *SetParams) Getparams() domain.Options {
	return p.Service.GetParams()
}

func checkInput(opt domain.Options) error {
	if opt.Side != "buy" && opt.Side != "sell" && opt.Side != "" {
		return errors.New(`'side' option must be 'buy' or 'sell'`)
	}
	if opt.Size < 1 {
		return errors.New(`'size' option must be more than 0`)
	}
	if opt.Profit <= 0 {
		return errors.New(`'profit' must be more than 0`)
	}
	if opt.Ticker != "PI_XBTUSD" && opt.Ticker != "PI_ETHUSD" && opt.Ticker != "PI_LTCUSD" && opt.Ticker != "PI_XRPUSD" && opt.Ticker != "PI_BCHUSD" {
		return errors.New(`'ticker' option must be 'PI_XBTUSD' or 'PI_ETHUSD' or 'PI_LTCUSD' or 'PI_XRPUSD' or 'PI_BCHUSD'`)
	}
	return nil
}

func checkInputWithStart(opt domain.Options) error {
	if opt.Start != 0 && opt.Start != 1 {
		return errors.New(`'start' option must be '1' or '0'`)
	}
	err := checkInput(opt)
	if err != nil {
		return err
	}
	return nil
}
