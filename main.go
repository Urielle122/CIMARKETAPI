// main.go
package main

import (
	"log"
	"net/http"
	"os"

	"services/core"
	logs "services/log"
	"services/routes"

	"github.com/joho/godotenv"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		logs.InfoF("CORS headers set for %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func main() {
	// ✅ Charger les variables d'environnement depuis le fichier .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Fichier .env non trouvé, utilisation des variables d'environnement système")
	} else {
		log.Println("✅ Fichier .env chargé avec succès")
	}

	core.InitConnection()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", routes.AddContactWithTransaction)

	handlerWithCORS := enableCORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback pour le développement local
	}

	log.Println("🚀 Serveur démarré sur le port", port)
	log.Fatal(http.ListenAndServe(":"+port, handlerWithCORS))
}
