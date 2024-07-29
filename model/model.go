package model

type Message struct {
	ID        int    `json:"id,omitempty" pg:",pk,notnull"`
	Content   string `json:"content,omitempty" pg:",notnull"`
	From      string `json:"from,omitempty" pg:",notnull"`
	To        string `json:"to,omitempty" pg:",notnull"`
	Timestamp int64  `json:"timestamp,omitempty" pg:",notnull,default:extract(epoch from now())"`
	Status    string `json:"status,omitempty" pg:",notnull,default:'new'"`
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
