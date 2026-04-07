package pipeline

import "time"

type Document struct {
	ID       string
	Title    string
	Path     string
	Content  string
	Size     int64
	Modified time.Time
	Type     string
	Tokens   []string
}
