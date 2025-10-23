// routes/contact.go
package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"services/core"
	"services/email"
	logs "services/log"
	"services/models"
	"time"
)

func AddContactWithTransaction(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Success bool                  `json:"success"`
		Message string                `json:"message"`
		Data    *models.ContactModels `json:"data,omitempty"`
	}

	var body models.ContactModels
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logs.Errorf("Erreur lors du décodage du JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Format de données invalide",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := core.MysqlDb.BeginTx(ctx, nil)
	if err != nil {
		logs.Errorf("Erreur lors du démarrage de la transaction: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Erreur interne du serveur",
		})
		return
	}

	query := `INSERT INTO contact_form (nom, email, phone, formation, message) VALUES (?, ?, ?, ?, ?)`

	res, err := tx.ExecContext(ctx, query, body.Nom, body.Email, body.Phone, body.Formation, body.Message)
	if err != nil {
		tx.Rollback()
		logs.Errorf("Erreur lors de l'insertion: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Erreur lors de l'ajout du contact",
		})
		return
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		logs.Errorf("Erreur lors de la récupération de l'ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Erreur lors de l'ajout du contact",
		})
		return
	}

	insertedContact := models.ContactModels{
		ID:        int(lastID),
		Nom:       body.Nom,
		Email:     body.Email,
		Phone:     body.Phone,
		Formation: body.Formation,
		Message:   body.Message,
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("Erreur lors du commit: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Success: false,
			Message: "Erreur lors de la finalisation de l'ajout",
		})
		return
	}

	logs.Info("Transaction commitée avec succès")

	// ✅ CHARGER LA CONFIGURATION DEPUIS LES VARIABLES D'ENVIRONNEMENT
	emailConfig := email.LoadEmailConfig()
	recipientEmail := email.GetRecipientEmail()

	// Vérifier que l'email du destinataire est configuré
	if recipientEmail == "" {
		logs.Errorf("Email du destinataire non configuré dans les variables d'environnement")
	} else {
		// Envoyer l'email de manière asynchrone
		go func() {
			if err := email.SendContactEmail(insertedContact, recipientEmail, emailConfig); err != nil {
				logs.Errorf("Erreur lors de l'envoi de l'email: %v", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: "Contact ajouté avec succès et email envoyé",
		Data:    &insertedContact,
	})
}
