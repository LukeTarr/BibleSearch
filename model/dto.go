package model

type QueryDTO struct {
	Query string `json:"query"`
}

type Metadata struct {
	Book    string `json:"book"`
	Chapter string `json:"chapter"`
	Verse   string `json:"verse"`
}

type QueryResultsDTO struct {
	Metadata Metadata `json:"metadata"`
	Distance float64  `json:"distance"`
	Text     string   `json:"text"`
	Id       string   `json:"id"`
}

type StatusDTO struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorDTO struct {
	Error string `json:"error"`
}
