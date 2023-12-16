package model

type QueryDTO struct {
	Query string `json:"query"`
}

type VectorizeDTO struct {
	Password string `json:"password"`
}

type Metadata struct {
	Book    string `json:"book"`
	Chapter string `json:"chapter"`
	Verse   string `json:"verse"`
}

type ChromaQueryResultsDTO struct {
	Metadata Metadata `json:"metadata"`
	Distance float64  `json:"distance"`
	Text     string   `json:"text"`
	Id       string   `json:"id"`
}

type QueryResultsDTO struct {
	Result []ChromaQueryResultsDTO `json:"result"`
}

type StatusDTO struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorDTO struct {
	Error string `json:"error"`
}
