CREATE TABLE IF NOT EXISTS athletes (
    id VARCHAR(255) PRIMARY KEY,
    nom VARCHAR(255),
    prenom VARCHAR(50),
    age VARCHAR(10) DEFAULT NULL
);


CREATE TABLE IF NOT EXISTS contact_form (
    id INT PRIMARY KEY AUTO INCREMENT,
    nom VARCHAR(255),
    email VARCHAR(250),
    phone VARCHAR(20),
    formation VARCHAR(250),
    message VARCHAR(2000)
);


-- Table des documents/pièces jointes des athlètes
CREATE TABLE IF NOT EXISTS documents_athletes (
    id VARCHAR(255) PRIMARY KEY,
    athlete_id VARCHAR(255) NOT NULL,
    type_document VARCHAR(50) NOT NULL,  -- 'PASSEPORT', 'CNI', 'LICENCE', etc.
    nom_fichier VARCHAR(255) NOT NULL,
    chemin_fichier VARCHAR(500) NOT NULL,
    FOREIGN KEY (athlete_id) REFERENCES athletes(id) ON DELETE CASCADE
);

-- Index pour optimiser la recherche des documents par athlète
CREATE INDEX idx_athlete_docs ON documents_athletes(athlete_id);
CREATE INDEX idx_type_document ON documents_athletes(type_document);


CREATE TABLE athletes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nom VARCHAR(255),
    prenom VARCHAR(255),
    age INT
);

CREATE TABLE IF NOT EXISTS documents_athletes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),  
    athlete_id UUID NOT NULL,
    type_document VARCHAR(50) NOT NULL,  
    nom_fichier VARCHAR(255) NOT NULL,
    chemin_fichier VARCHAR(500) NOT NULL,
    FOREIGN KEY (athlete_id) REFERENCES athletes(id) ON DELETE CASCADE
);

CREATE INDEX idx_athlete_docs ON documents_athletes(athlete_id);
CREATE INDEX idx_type_document ON documents_athletes(type_document);
CREATE INDEX idx_athlete_type_docs ON documents_athletes(athlete_id, type_document);       
