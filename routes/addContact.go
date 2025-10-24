package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"services/core"
	"services/email"
	logs "services/log"
	"services/models"
)

func sendEmailWithFallback(nom, emailAddr, phone, formation, message string) error {
	contact := models.ContactModels{
		Nom:       nom,
		Email:     emailAddr,
		Phone:     phone,
		Formation: formation,
		Message:   message,
	}

	// 1. Essayer Resend
	err := email.SendResendEmail(contact)
	if err == nil {
		return nil // ✅ Resend a marché
	}

	logs.Warnf("Resend a échoué: %v - Essai avec Gmail...", err)

	// 2. Essayer Gmail en backup
	config := email.LoadEmailConfig()
	toEmail := email.GetRecipientEmail()
	if toEmail == "" {
		toEmail = emailAddr
	}

	err = email.SendContactEmail(contact, toEmail, config)
	if err == nil {
		return nil // ✅ Gmail a marché
	}

	return fmt.Errorf("les deux méthodes ont échoué")
}

// Ajoute cette fonction dans routes/addContact.go
// func sendContactEmail(nom, emailAddr, phone, formation, message string) error {
// 	// Crée l'objet ContactModels
// 	contact := models.ContactModels{
// 		Nom:       nom,
// 		Email:     emailAddr,
// 		Phone:     phone,
// 		Formation: formation,
// 		Message:   message,
// 	}

// 	// Charge la configuration email
// 	config := email.LoadEmailConfig()

// 	// Récupère l'email du destinataire
// 	toEmail := email.GetRecipientEmail()
// 	if toEmail == "" {
// 		toEmail = emailAddr // Fallback vers l'email du contact
// 	}

// 	// Appelle la fonction avec les bons arguments
// 	return email.SendContactEmail(contact, toEmail, config)
// }

func AddContactWithTransaction(w http.ResponseWriter, r *http.Request) {
	// ✅ Récupère la connexion DB
	db := core.GetDB()
	if db == nil {
		logs.Error("Database connection is nil")
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Lit le corps de la requête
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logs.Error("Error reading request body", "error", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse les données JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		logs.Error("Error parsing JSON", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extraction des données
	nom, _ := data["nom"].(string)
	emailAddr, _ := data["email"].(string)
	phone, _ := data["phone"].(string)
	formation, _ := data["formation"].(string)
	message, _ := data["message"].(string)

	// Début de la transaction
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logs.Error("Error starting transaction", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insertion dans la base
	query := "INSERT INTO contact_form (nom, email, phone, formation, message) VALUES (?, ?, ?, ?, ?)"
	result, err := tx.ExecContext(ctx, query, nom, emailAddr, phone, formation, message)
	if err != nil {
		logs.Error("Error inserting contact", "error", err)
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	// Commit de la transaction
	if err := tx.Commit(); err != nil {
		logs.Error("Error committing transaction", "error", err)
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	// Récupère l'ID inséré
	lastID, _ := result.LastInsertId()
	logs.Info("Transaction committed successfully", "contact_id", lastID)

	// Envoi de l'email
	if err := sendEmailWithFallback(nom, emailAddr, phone, formation, message); err != nil {
		logs.Warnf("Email non envoyé mais contact sauvegardé", "error", err)
	}

	// Réponse de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Contact added successfully",
		"id":      lastID,
	})
}
