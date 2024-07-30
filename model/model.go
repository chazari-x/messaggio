package model

type Message struct {
	ID        int    `json:"id" pg:",pk,notnull"`
	Content   string `json:"content" pg:",notnull"`
	From      string `json:"from" pg:",notnull"`
	To        string `json:"to" pg:",notnull"`
	Timestamp int64  `json:"timestamp" pg:",notnull,default:extract(epoch from now())"`
	Status    string `json:"status" pg:",notnull,default:'new'"`
}

type Status string

const (
	New        Status = "new"
	Processing Status = "processing"
	Ok         Status = "ok"
	Error      Status = "error"
)

func (s Status) String() string {
	return string(s)
}
