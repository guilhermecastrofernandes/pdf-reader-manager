package main

import (
	"fmt"
	"log"
	"os"

	"your_project/api"
	"your_project/models"
	"your_project/pdf"

	"github.com/joho/godotenv"
)

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	pdfPath := os.Getenv("PDF_PATH")
	cards, err := processPdf(pdfPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, card := range cards {
		fmt.Println("Card Details:")
		fmt.Printf("  Date:  %s\n", card.Date)
		fmt.Printf("  Store: %s\n", card.Store)
		fmt.Printf("  Value: %s\n", card.Value)
		fmt.Printf("  Name:  %s\n", card.Name)

		err := api.MakePostRequest(card)
		if err != nil {
			log.Printf("Error sending card data to API: %v", err)
		}
	}
}

func loadEnv() error {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Failed to load file .env:", err)
		return err
	}
	return nil
}

func processPdf(path string) ([]models.Card, error) {
	reader, err := pdf.LoadPDF(path)
	if err != nil {
		return nil, err
	}

	cards, err := pdf.ParseCardData(reader)
	if err != nil {
		return nil, err
	}

	return cards, nil
}
