package models

type FormData struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
}

type EmailTemplateData struct {
	FormData      FormData
	SubmittedTime string
	SubmittedDate string
	Origin        string
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
