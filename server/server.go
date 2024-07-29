package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"messaggio/broker"
	"messaggio/config"
	_ "messaggio/docs"
	"messaggio/model"
	"messaggio/storage"
)

type Server struct {
	cfg     config.Server
	storage *storage.Storage
	kafka   *broker.Broker
	server  *http.Server
}

func New(cfg config.Server, s *storage.Storage, k *broker.Broker) *Server {
	return &Server{
		cfg:     cfg,
		storage: s,
		kafka:   k,
	}
}

func (s *Server) Start() {
	go func() {
		r := chi.NewRouter()
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		r.Post("/api/messages", s.createMessage)
		r.Get("/api/messages", s.getMessages)
		r.Get("/api/messages/{id}", s.getMessage)

		s.server = &http.Server{
			Addr:    s.cfg.Http,
			Handler: r,
		}

		log.Info("http server starting..")
		if err := s.server.ListenAndServe(); err != nil {
			log.Error(err)
		}
		log.Info("http server stopped")
	}()
}

func (s *Server) Close(ctx context.Context) {
	_ = s.server.Shutdown(ctx)
}

type responseMessages struct {
	Status   string          `json:"status"`
	Messages []model.Message `json:"messages,omitempty"`
}

func (r responseMessages) Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	marshal, _ := json.Marshal(r)
	_, _ = w.Write(marshal)
}

type responseMessage struct {
	Status  string        `json:"status"`
	Message model.Message `json:"message,omitempty"`
}

func (r responseMessage) Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	marshal, _ := json.Marshal(r)
	_, _ = w.Write(marshal)
}

type responseError struct {
	Status string `json:"status"`
	Text   string `json:"text,omitempty"`
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
