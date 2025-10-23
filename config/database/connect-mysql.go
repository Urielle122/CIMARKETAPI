package database

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"
	logs "services/log"

	"github.com/go-sql-driver/mysql" // Supprimez le _ et utilisez le package directement
)

func ConnectToMySQL() (*sql.DB, error) {
	caCert, err := os.ReadFile("ssl/ca.pem")
	if err != nil {
		return nil, fmt.Errorf("erreur lecture certificat: %v", err)
	}

	// 2. Créer le pool de certificats
	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("erreur chargement certificat CA")
	}

	// 3. Configurer TLS
	tlsConfig := &tls.Config{
		RootCAs: rootCertPool,
	}

	// 4. Enregistrer la config TLS
	err = mysql.RegisterTLSConfig("aiven", tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("erreur config TLS: %v", err)
	}

	// Paramètres de connexion
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if port == "" {
		port = "3306"
	}

	// Construction DSN avec paramètres SSL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true&timeout=30s",
		user, password, host, port, dbname)

	// Ouverture de la connexion
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logs.Errorf("Erreur lors de l'ouverture de la connexion mySql: %v", err)
		return nil, err
	}

	// Test de la connexion
	err = db.Ping()
	if err != nil {
		logs.Errorf("Erreur lors du test de connexion à MYSQL : %v", err)
		return nil, err
	}

	logs.Info("Connecté avec succès à MYSQL")
	return db, nil
}

// Fonction optionnelle pour fermer la connexion
func CloseConnection(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			logs.Errorf("Erreur lors de la fermeture de la connexion : %v", err)
		}
	}
}
