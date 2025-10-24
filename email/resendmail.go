package email

import (
	"fmt"
	logs "services/log"
	"services/models"

	"github.com/resend/resend-go/v2"
)

func SendResendEmail(contact models.ContactModels) error {
	// 1. Récupérer la clé Resend
	apiKey := getEnv("RESEND_API_KEY", "")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY non configurée")
	}

	// 2. Créer le client Resend
	client := resend.NewClient(apiKey)

	// 3. Email du destinataire
	toEmail := GetRecipientEmail()
	if toEmail == "" {
		toEmail = "urielle.age@gmail.com" // Ton email
	}

	// 4. Créer le contenu HTML
	htmlContent := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px;">
			<h2 style="color: #f97316;">🎉 Nouveau Contact CI Marketing</h2>
			<div style="background: #f8f9fa; padding: 15px; border-radius: 5px;">
				<h3>Informations du contact :</h3>
				<p><strong>Nom :</strong> %s</p>
				<p><strong>Email :</strong> %s</p>
				<p><strong>Téléphone :</strong> %s</p>
				<p><strong>Formation :</strong> %s</p>
				<p><strong>Message :</strong> %s</p>
			</div>
		</div>
	`, contact.Nom, contact.Email, contact.Phone, contact.Formation, contact.Message)

	// 5. Envoyer l'email
	params := &resend.SendEmailRequest{
		From:    getEnv("RESEND_FROM_EMAIL", "CI Marketing <onboarding@resend.dev>"),
		To:      []string{toEmail},
		Subject: fmt.Sprintf("📧 Nouveau contact: %s", contact.Nom),
		Html:    htmlContent,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		logs.Errorf("❌ Erreur Resend: %v", err)
		return err
	}

	logs.InfoF("✅ Email Resend envoyé! ID: %s", sent.Id)
	return nil
}
