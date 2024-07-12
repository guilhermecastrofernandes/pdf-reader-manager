package domain

type RequestData struct {
	Date  string `json:"date"`
	Store string `json:"store"`
	Value string `json:"value"`
	Name  string `json:"name"`
}
