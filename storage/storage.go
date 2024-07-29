package storage

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"messaggio/config"
	"messaggio/model"
)

type Storage struct {
	db  *pg.DB
	cfg config.DataBase
}

func New(ctx context.Context, cfg config.DataBase) (*Storage, error) {
	opt, err := pg.ParseURL(cfg.Addr)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(opt)

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	s := &Storage{
		cfg: cfg,
		db:  db,
	}

	return s, s.Create()
}

func (s *Storage) Close() {
	_ = s.db.Close()
}

func (s *Storage) Create() error {
	return s.db.Model((*model.Message)(nil)).CreateTable(&orm.CreateTableOptions{
		IfNotExists: true,
	})
}

func (s *Storage) Insert(msg *model.Message) error {
	_, err := s.db.Model(msg).Returning("id, status, timestamp").Insert()
	return err
}

func (s *Storage) SelectAll(status string, page, limit int) ([]model.Message, error) {
	var msgs []model.Message
	query := s.db.Model(&msgs)
	if status != "" {
		query.Where("status = ?", status)
	}
	query.Offset((page - 1) * limit).Limit(limit).Order("id")
	err := query.Select()
	return msgs, err
}

func (s *Storage) SelectById(id int) (model.Message, error) {
	var msg model.Message
	err := s.db.Model(&msg).Where("id = ?", id).Select()
	return msg, err
}

func (s *Storage) SelectNew() ([]model.Message, error) {
	var msgs []model.Message
	err := s.db.Model(&msgs).Where("status = ?", model.New).Order("id").Select()
	return msgs, err
}

func (s *Storage) UpdateStatus(id int, status model.Status) error {
	_, err := s.db.Model(&model.Message{ID: id}).Set("status = ?", status).WherePK().Update()
	return err
}

func (s *Storage) UpdateStatuses(msgs []model.Message, status model.Status) error {
	_, err := s.db.Model(&msgs).Set("status = ?", status).WherePK().Update()
	return err
}
