package models

type ContactModels struct {
	ID        int    `json:"id"`
	Nom       string `json:"nom"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Formation string `json:"formation"`
	Message   string `json:"message"`
}
