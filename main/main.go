package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/unidoc/unipdf/v3/model"
)

type Card struct {
	Date  string
	Store string
	Value string
	Name  string
}

type RequestData struct {
	Date  string `json:"date"`
	Store string `json:"store"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

func main() {

	loadEnv()
	readPdf()

}

func loadEnv() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Failed to load file .env:", err)
	}
}

func readPdf() {
	// Open PDF file.

	cards := []Card{}
	card := Card{}
	name := ""

	pdfFile, err := os.Open(os.Getenv("PDF_PATH"))
	if err != nil {
		fmt.Println("Failed to open file:", err)

	}

	// Create a new reader of page.
	reader, err := model.NewPdfReader(pdfFile)

	if err != nil {
		fmt.Println("Failed to create a pdf file:", err)

	}

	// Get number of pages
	numPages, err := reader.GetNumPages()
	if err != nil {
		fmt.Println("Failed to get number of pages:", err)

	}
	for i := 2; i <= numPages; i++ {
		// Get main Page
		page, err := reader.GetPage(i)
		if err != nil {
			fmt.Println("Failed to get main page", err)
			continue
		}

		//Get text from page.
		lines, err := page.GetContentStreams()
		if err != nil {
			fmt.Println("Failed to extract text from page:", err)
			continue
		}

		regexData := regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])/(0[1-9]|1[0-2])$`)

		arrayCount := 0

		for _, line := range lines {
			var buffer bytes.Buffer

			re := regexp.MustCompile(`\([^)]+\)`)
			matches := re.FindAllString(line, -1)
			for _, match := range matches {
				extrated := match[1 : len(match)-1] // Remove parentheses due there are parentheses when read a pdf file.
				fmt.Fprintln(&buffer, extrated)

			}
			lineWithFilter := buffer.String()
			lineIntoArray := strings.Split(lineWithFilter, "\n")

			for _, line := range lineIntoArray {
				array := lineIntoArray
				arrayCount++

				if haveCard(line) {
					name = line

				}

				if regexData.MatchString(line) {

					position := arrayCount

					store := array[position]
					value := array[position+1]
					card.Store = store
					card.Value = value
					card.Date = line
					card.Name = name
					cards = append(cards, card)

				}
			}

		}

	}

	for _, currentCard := range cards {
		fmt.Printf("Card Details:\n  Date:  %s\n  Store: %s\n  Value: %s\n  Name:  %s\n", currentCard.Date, currentCard.Store, currentCard.Value, currentCard.Name)
		requestData := convertCardToRequestData(currentCard)
		makePostRequest(requestData)
	}

}

func haveCard(line string) bool {
	cards := []string{os.Getenv("FIRST_USER_CARD"), os.Getenv("FIRST_USER_VIRTUAL_CARD"),
		os.Getenv("SECOND_USER_CARD"), os.Getenv("SECOND_USER_VIRTUAL_CARD")}

	for _, card := range cards {
		if strings.Contains(line, card) {
			return true
		}
	}
	return false
}

func makePostRequest(data RequestData) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", os.Getenv("SHEET_DB_URL"), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("LOGIN"), os.Getenv("PASSWORD"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erro ao chamar API", err)
		return nil, err
	}

	return resp, nil
}

func convertCardToRequestData(card Card) RequestData {
	return RequestData{
		Date:  card.Date,
		Store: card.Store,
		Value: card.Value,
		Name:  card.Name,
	}
}
