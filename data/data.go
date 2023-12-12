package data

import (
	"BibleSearch/model"
	"encoding/json"
	"os"
	"strings"
)

func GetBookSlice() (*[]model.Book, error) {
	bible, err := os.ReadFile("./data/en_kjv.json")
	if err != nil {
		return nil, err
	}
	var books []model.Book

	// Remove BOMs from the beginning of the file
	stringBible := string(bible)
	jsonData := strings.TrimPrefix(stringBible, "\xef\xbb\xbf")

	if err := json.Unmarshal([]byte(jsonData), &books); err != nil {
		return nil, err
	}

	return &books, nil
}
