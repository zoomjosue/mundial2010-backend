package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "database/sql/driver"
)

var DB *sql.DB

func Connect() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		dsn = withSSLMode(dsn)
	} else {
		dsn = buildDSNFromParts()
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	DB = db
	return db, nil
}

func buildDSNFromParts() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "mundial2010")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func withSSLMode(dsn string) string {
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" || strings.Contains(dsn, "sslmode=") {
		return dsn
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}

	return dsn + separator + "sslmode=" + url.QueryEscape(sslmode)
}

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS series (
			id          SERIAL PRIMARY KEY,
			name        VARCHAR(255) NOT NULL,
			description TEXT,
			genre       VARCHAR(100),
			year        INTEGER,
			image_url   TEXT DEFAULT '',
			created_at  TIMESTAMP DEFAULT NOW(),
			updated_at  TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id         SERIAL PRIMARY KEY,
			serie_id   INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
			score      INTEGER NOT NULL CHECK (score >= 1 AND score <= 10),
			comment    TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ratings_serie_id ON ratings(serie_id)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}

	seedData(db)

	log.Println("Database migrated")
	return nil
}

func seedData(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM series").Scan(&count)
	if count > 0 {
		return
	}

	series := []struct {
		name, description, genre string
		year                     int
	}{
		{"España Campeona", "La historia épica de cómo España ganó su primer Mundial en Sudáfrica 2010, culminando con el gol de Andrés Iniesta en la final contra Países Bajos.", "Documental", 2010},
		{"La Roja - El Camino", "Recorrido partido a partido de la selección española durante el Mundial de Sudáfrica: desde el tropiezo inicial contra Suiza hasta el título.", "Serie Documental", 2010},
		{"Iker Casillas: El Capitán", "Documental sobre el portero y capitán de España, Iker Casillas, y su liderazgo durante el Mundial de Sudáfrica 2010.", "Biografía", 2010},
		{"Iniesta: El Gol Eterno", "La historia del gol en el minuto 116 de la prórroga que le dio a España su primera Copa del Mundo frente a Países Bajos.", "Documental", 2010},
		{"Villa - El Guaje", "David Villa, máximo goleador del torneo con 5 goles, protagoniza este documental sobre su torneo histórico en Sudáfrica.", "Biografía", 2010},
		{"Magia del Tiquitaca", "Análisis del estilo de juego revolucionario de la selección española que dominó el fútbol mundial en 2010.", "Análisis Deportivo", 2010},
		{"Del Bosque: El Arquitecto", "Vicente del Bosque y su visión táctica que llevó a España a conquistar el mundo con un fútbol de posesión sin precedentes.", "Documental", 2010},
		{"Sudáfrica 2010: El Tour", "Las 10 sedes del primer Mundial africano: Johannesburgo, Ciudad del Cabo, Durban y más, contadas con historia y cultura local.", "Viajes", 2010},
	}

	for _, s := range series {
		db.Exec(`INSERT INTO series (name, description, genre, year) VALUES ($1, $2, $3, $4)`,
			s.name, s.description, s.genre, s.year)
	}
	log.Println("Seed data inserted")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
