package model

type Book struct {
	Abbrev   string     `json:"abbrev"`
	Name     string     `json:"name"`
	Chapters [][]string `json:"chapters"`
}
