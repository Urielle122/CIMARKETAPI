// services/email/email.go
package email

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	logs "services/log"
	"services/models"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// Charger la configuration depuis les variables d'environnement
func LoadEmailConfig() EmailConfig {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		port = 587 // Port par défaut
	}

	return EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     port,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("SMTP_FROM_EMAIL", ""),
		FromName:     getEnv("SMTP_FROM_NAME", "Formulaire de Contact"),
	}
}

// Fonction helper pour récupérer les variables d'environnement avec valeur par défaut
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Récupérer l'email du destinataire
func GetRecipientEmail() string {
	return getEnv("RECIPIENT_EMAIL", "")
}

func generateCSV(contact models.ContactModels) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// En-têtes
	headers := []string{"ID", "Nom", "Email", "Téléphone", "Formation", "Message"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Données
	record := []string{
		fmt.Sprintf("%d", contact.ID),
		contact.Nom,
		contact.Email,
		contact.Phone,
		contact.Formation,
		contact.Message,
	}
	if err := writer.Write(record); err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func SendContactEmail(contact models.ContactModels, toEmail string, config EmailConfig) error {
	// Générer le CSV
	csvData, err := generateCSV(contact)
	if err != nil {
		logs.Errorf("Erreur lors de la génération du CSV: %v", err)
		return err
	}

	// Créer le message
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(config.FromEmail, config.FromName))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Nouveau contact: %s", contact.Nom))

	// Corps de l'email en HTML
	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>Nouveau contact reçu</h2>
            <p><strong>Nom:</strong> %s</p>
            <p><strong>Email:</strong> %s</p>
            <p><strong>Téléphone:</strong> %s</p>
            <p><strong>Formation:</strong> %s</p>
            <p><strong>Message:</strong></p>
            <p>%s</p>
            <hr>
            <p><em>Veuillez trouver ci-joint le fichier CSV avec les détails complets.</em></p>
        </body>
        </html>
    `, contact.Nom, contact.Email, contact.Phone, contact.Formation, contact.Message)

	m.SetBody("text/html", body)

	// Ajouter le CSV en pièce jointe
	filename := fmt.Sprintf("contact_%d.csv", contact.ID)
	m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(csvData)
		return err
	}))

	// Configurer le dialer
	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)

	// Envoyer l'email
	if err := d.DialAndSend(m); err != nil {
		logs.Errorf("Erreur lors de l'envoi de l'email: %v", err)
		return err
	}

	logs.Info("Email envoyé avec succès à:", toEmail)
	return nil
}
