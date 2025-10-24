package database

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"
	logs "services/log"

	"github.com/go-sql-driver/mysql"
)

func ConnectToMySQL() (*sql.DB, error) {
	logs.Info("🔧 Début connexion MySQL...")

	// 1. Essayer de lire le certificat depuis SQL_CERTIFICATE
	sqlCert := os.Getenv("SQL_CERTIFICATE")
	var caCert []byte
	var err error

	if sqlCert != "" {
		logs.Info("📁 Certificat trouvé dans SQL_CERTIFICATE")

		// ✅ CORRECTION : Utiliser directement comme texte, pas de base64
		caCert = []byte(sqlCert)
		logs.Info("✅ Certificat chargé directement (format texte)")
	} else {
		logs.Warnf("⚠️  SQL_CERTIFICATE non trouvé, tentative sans certificat")
		return connectWithoutCert()
	}

	// 2. Créer le pool de certificats
	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(caCert); !ok {
		logs.Error("❌ Erreur chargement certificat CA - le format est peut-être incorrect")
		// Afficher un extrait pour debug
		if len(caCert) > 100 {
			logs.InfoF("Extrait certificat: %s...", string(caCert[:100]))
		}
		return connectWithoutCert() // Fallback
	}

	// 3. Configurer TLS
	tlsConfig := &tls.Config{
		RootCAs: rootCertPool,
	}

	// 4. Enregistrer la config TLS
	err = mysql.RegisterTLSConfig("aiven", tlsConfig)
	if err != nil {
		logs.Errorf("❌ Erreur config TLS: %v", err)
		return connectWithoutCert() // Fallback
	}

	logs.Info("🔐 Configuration TLS avec certificat réussie")

	// Paramètres de connexion
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if port == "" {
		port = "3306"
	}

	logs.InfoF("🔗 Connexion à: %s@%s:%s", user, host, port)

	// Construction DSN avec certificat
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true&timeout=30s",
		user, password, host, port, dbname)

	// Ouverture de la connexion
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logs.Errorf("❌ Erreur ouverture connexion: %v", err)
		return nil, err
	}

	// Test de la connexion
	err = db.Ping()
	if err != nil {
		logs.Errorf("❌ Erreur test connexion: %v", err)
		return nil, err
	}

	logs.Info("✅ Connecté avec succès à MYSQL (avec certificat)")
	return db, nil
}

// ✅ FONCTION DE FALLBACK AVEC TLS SKIP VERIFY (pour Aiven)
func connectWithoutCert() (*sql.DB, error) {
	logs.Info("🔄 Tentative de connexion avec TLS skip verify...")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if port == "" {
		port = "3306"
	}

	logs.InfoF("🔗 Connexion à: %s@%s:%s", user, host, port)

	// ✅ Configuration TLS qui accepte les certificats auto-signés
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // ⚠️ Pour Aiven en développement
	}

	mysql.RegisterTLSConfig("skipverify", tlsConfig)

	// Utilise TLS avec skip verify
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skipverify&parseTime=true&timeout=30s",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logs.Errorf("❌ Erreur ouverture connexion: %v", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logs.Errorf("❌ Erreur test connexion: %v", err)
		return nil, err
	}

	logs.Info("✅ Connecté avec succès à MYSQL (TLS skip verify)")
	return db, nil
}

// Fonction optionnelle pour fermer la connexion
func CloseConnection(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			logs.Errorf("Erreur fermeture connexion: %v", err)
		}
	}
}
