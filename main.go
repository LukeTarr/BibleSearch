package main

import (
	"fmt"
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/joho/godotenv"
	"os"
)

func main() {

	fmt.Println("Starting BibleSearch Engine")
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading .env file: %s", err)
		return
	}

	fmt.Println("Building Chroma Client")
	client := chroma.NewClient(os.Getenv("CHROMA_URL"))
	meta := map[string]interface{}{}

	fmt.Println("Creating Embedding Function")
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(os.Getenv("OPENAI_API_KEY")) //create a new OpenAI Embedding function

	fmt.Println("Creating Collection")
	collection, _ := client.CreateCollection("test", meta, true, embeddingFunction, chroma.L2)

	fmt.Println("Adding Documents, ids, and metadatas")
	documents := []string{
		"This is a document about cats. Cats are great.",
		"this is a document about dogs. Dogs are great.",
	}
	ids := []string{
		"ID1",
		"ID2",
	}

	metadatas := []map[string]interface{}{
		{"key1": "value1"},
		{"key2": "value2"},
	}

	fmt.Println("Adding Documents to Collection")
	col, addError := collection.Add(nil, metadatas, documents, ids)
	if addError != nil {
		fmt.Printf("Error adding documents: %s", addError)
		return
	}

	fmt.Printf("col: %v\n", col) //this should result in the collection with the two documents

	countDocs, qrerr := collection.Count()
	if qrerr != nil {
		fmt.Printf("Error counting documents: %s", qrerr)
		return
	}
	fmt.Printf("countDocs: %v\n", countDocs) //this should result in 2
	qr, qrerr := collection.Query([]string{"I love dogs"}, 5, nil, nil, nil)
	if qrerr != nil {
		fmt.Printf("Error `querying documents: %s", qrerr)
		return
	}
	fmt.Printf("qr: %v\n", qr.Documents[0][0]) //this should result in the document about dogs
}
