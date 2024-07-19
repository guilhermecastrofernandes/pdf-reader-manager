package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"pdf-reader/domain"

	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/unidoc/unipdf/v3/model"
)

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

	card := domain.Card{}
	cards := []domain.Card{}
	newCards := append(cards, card)
	name := ""

	files := loadPDF()

	for _, file := range files {
		pdfFile, err := os.Open(file)
		if err != nil {
			fmt.Println("Failed to open file:", err)

		}

		reader, err := model.NewPdfReader(pdfFile)
		if err != nil {
			fmt.Println("Failed to create a reader pdf :", err)

		}

		numPages, err := reader.GetNumPages()
		if err != nil {
			fmt.Println("Failed to get number of pages:", err)

		}
		for i := 2; i <= numPages; i++ {
			mainPage, err := reader.GetPage(i)
			if err != nil {
				fmt.Println("Failed to get main page", err)
				continue
			}

			//Get text from page.
			lines, err := mainPage.GetContentStreams()
			if err != nil {
				fmt.Println("Failed to extract text from page:", err)
				continue
			}

			regexDate := regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])/(0[1-9]|1[0-2])$`)

			arrayCount := 0

			for _, line := range lines {
				var buffer bytes.Buffer

				re := regexp.MustCompile(`\([^)]+\)`)
				matches := re.FindAllString(line, -1)
				for _, match := range matches {
					// Remove parentheses due there are parentheses when read a pdf file.
					extratedWithoutParentheses := match[1 : len(match)-1]
					fmt.Fprintln(&buffer, extratedWithoutParentheses)

				}

				lineWithFilter := buffer.String()
				lineIntoArray := strings.Split(lineWithFilter, "\n")

				for _, line := range lineIntoArray {
					array := lineIntoArray
					arrayCount++

					if haveCardInformation(line) {
						name = line

					}

					if regexDate.MatchString(line) {

						position := arrayCount

						store := array[position]
						value := array[position+1]
						card.Store = store
						card.Value = value
						card.Date = line
						card.Name = name
						newCards = append(newCards, card)

					}
				}

			}

		}

	}

	for _, currentCard := range newCards {
		fmt.Printf("Card Details:\n  Date:  %s\n  Store: %s\n  Value: %s\n  Name:  %s\n", currentCard.Date, currentCard.Store, currentCard.Value, currentCard.Name)
		requestData := convertCardToRequestData(currentCard)
		callSheetAPI(requestData)
	}

}

func loadPDF() []string {
	pdfDirPath := os.Getenv("PDF_PATH")

	if _, err := os.Stat(pdfDirPath); err != nil {
		log.Fatal(err)
	}

	var paths []string

	err := filepath.Walk(pdfDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fmt.Println("Processando arquivo:", path)
		paths = append(paths, path)

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	return paths
}

func haveCardInformation(line string) bool {
	cards := []string{os.Getenv("FIRST_USER_CARD"), os.Getenv("FIRST_USER_VIRTUAL_CARD"),
		os.Getenv("SECOND_USER_CARD"), os.Getenv("SECOND_USER_VIRTUAL_CARD")}

	for _, card := range cards {
		if strings.Contains(line, card) {
			return true
		}
	}
	return false
}

func callSheetAPI(data domain.RequestData) (*http.Response, error) {
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

func convertCardToRequestData(card domain.Card) domain.RequestData {
	return domain.RequestData{
		Date:  card.Date,
		Store: card.Store,
		Value: card.Value,
		Name:  card.Name,
	}
}
