package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"messaggio/config"
	_ "messaggio/docs"
	"messaggio/model"
	"messaggio/prometheus"
	"messaggio/storage"
)

type Server struct {
	cfg        config.Server
	storage    *storage.Storage
	server     *http.Server
	prometheus *prometheus.Prometheus
}

func New(cfg config.Server, s *storage.Storage, p *prometheus.Prometheus) *Server {
	return &Server{
		cfg:        cfg,
		storage:    s,
		prometheus: p,
	}
}

func (s *Server) Start() {
	go func() {
		r := chi.NewRouter()

		r.Handle("/metrics", promhttp.Handler())

		r.Get("/api/swagger/*", httpSwagger.WrapHandler)

		r.Put("/api/messages/ok/add/{num}", s.AddOk)

		r.Put("/api/messages/new/add/{num}", s.AddNew)
		r.Put("/api/messages/new/sub/{num}", s.SubNew)

		r.Put("/api/messages/error/add/{num}", s.AddError)

		r.Put("/api/messages/processing/add/{num}", s.AddProcessing)
		r.Put("/api/messages/processing/sub/{num}", s.SubProcessing)

		r.Post("/api/messages", s.createMessage)
		r.Get("/api/messages", s.getMessages)
		r.Get("/api/messages/{id}", s.getMessage)

		s.server = &http.Server{
			Addr:    s.cfg.Http,
			Handler: r,
		}

		log.Infof("http server starting on %s", s.cfg.Http)
		if err := s.server.ListenAndServe(); err != nil {
			log.Error(err)
		}
		log.Info("http server stopped")
	}()
}

func (s *Server) Close(ctx context.Context) {
	_ = s.server.Shutdown(ctx)
}

func (s *Server) AddOk(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.OkMessageCounter.Add(float64(num))
}

func (s *Server) AddNew(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.NewMessageGauge.Add(float64(num))
}

func (s *Server) SubNew(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.NewMessageGauge.Sub(float64(num))
}

func (s *Server) AddError(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.ErrorMessageCounter.Add(float64(num))
}

func (s *Server) AddProcessing(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.ProcessingMessageGauge.Add(float64(num))
}

func (s *Server) SubProcessing(_ http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(chi.URLParam(r, "num"))
	if err != nil {
		return
	}

	s.prometheus.ProcessingMessageGauge.Sub(float64(num))
}

// @title Messaggio API
// @version 1.0
// @description This is a simple message broker
// @host localhost:8080
// @BasePath /api

type responseMessages struct {
	Status   string          `json:"status"`
	Messages []model.Message `json:"messages"`
}

func (r responseMessages) Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	marshal, _ := json.Marshal(r)
	_, _ = w.Write(marshal)
}

type responseMessage struct {
	Status  string        `json:"status"`
	Message model.Message `json:"message"`
}

func (r responseMessage) Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	marshal, _ := json.Marshal(r)
	_, _ = w.Write(marshal)
}

type responseError struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

func (r responseError) Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	marshal, _ := json.Marshal(r)
	_, _ = w.Write(marshal)
}

type request struct {
	Content string `json:"content"`
	From    string `json:"from"`
	To      string `json:"to"`
}

// @Summary Create a new message
// @Description Create a new message
// @Accept  json
// @Produce  json
// @Param   message  body  request  true  "Message content"
// @Success 201 {object} responseMessage
// @Failure 400 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/messages [post]
func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "произошла ошибка при чтении тела запроса",
		}.Write(w, http.StatusBadRequest)
		return
	}

	var reqMsg request
	if err = json.Unmarshal(body, &reqMsg); err != nil {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "неверный формат сообщения",
		}.Write(w, http.StatusBadRequest)
		return
	}

	var msg = model.Message{
		Content: reqMsg.Content,
		From:    reqMsg.From,
		To:      reqMsg.To,
	}
	if err = s.storage.Insert(&msg); err != nil {
		log.Error(err)
		responseError{
			Status: http.StatusText(http.StatusInternalServerError),
			Text:   "произошла ошибка при добавлении сообщения",
		}.Write(w, http.StatusInternalServerError)
		return
	}

	s.prometheus.NewMessageGauge.Inc()

	responseMessage{
		Status:  http.StatusText(http.StatusCreated),
		Message: msg,
	}.Write(w, http.StatusCreated)
}

// @Summary Get all messages
// @Description Get all messages
// @Produce  json
// @Param status query string false "Status filter" default("") Enums(new, processing, ok, error)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(50)
// @Success 200 {object} responseMessages
// @Failure 400 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/messages [get]
func (s *Server) getMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := r.URL.Query().Get("status")
	if !slices.Contains([]string{"", model.Ok.String(), model.New.String(), model.Processing.String(), model.Error.String()}, status) {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "неверный формат статуса",
		}.Write(w, http.StatusBadRequest)
		return
	}

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	intPage, err := strconv.Atoi(page)
	if err != nil {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "неверный формат страницы",
		}.Write(w, http.StatusBadRequest)
		return
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "неверный формат лимита",
		}.Write(w, http.StatusBadRequest)
		return
	}

	msgs, err := s.storage.SelectAll(status, intPage, intLimit)
	if err != nil {
		log.Error(err)
		responseError{
			Status: http.StatusText(http.StatusInternalServerError),
			Text:   "произошла ошибка при получении сообщений",
		}.Write(w, http.StatusInternalServerError)
		return
	}

	responseMessages{
		Status:   http.StatusText(http.StatusOK),
		Messages: msgs,
	}.Write(w, http.StatusOK)
}

// @Summary Get a message by ID
// @Description Get a message by ID
// @Produce  json
// @Param   id  path  string  true  "Message ID"
// @Success 200 {object} responseMessage
// @Failure 400 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/messages/{id} [get]
func (s *Server) getMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := chi.URLParam(r, "id")
	if id == "" {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "не указан id сообщения",
		}.Write(w, http.StatusBadRequest)
		return
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		responseError{
			Status: http.StatusText(http.StatusBadRequest),
			Text:   "неверный формат id сообщения",
		}.Write(w, http.StatusBadRequest)
		return
	}

	msg, err := s.storage.SelectById(intId)
	if err != nil {
		log.Error(err)
		responseError{
			Status: http.StatusText(http.StatusInternalServerError),
			Text:   "произошла ошибка при получении сообщения",
		}.Write(w, http.StatusInternalServerError)
		return
	}

	responseMessage{
		Status:  http.StatusText(http.StatusOK),
		Message: msg,
	}.Write(w, http.StatusOK)
}
