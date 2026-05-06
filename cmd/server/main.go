package main

import (
	"database/sql"
	"fmt"
	"log"
	"mundial2010/internal/db"
	"mundial2010/internal/handlers"
	"mundial2010/internal/middleware"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatal("Migration failed:", err)
	}

	seriesH := handlers.NewSeriesHandler(database)
	ratingsH := handlers.NewRatingsHandler(database)

	mux := http.NewServeMux()

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/swagger-ui.html", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","message":"Mundial 2010 API running"}`)
	})

	mux.HandleFunc("/series", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			seriesH.List(w, r)
		case http.MethodPost:
			seriesH.Create(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	mux.HandleFunc("/series/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/image") {
			if r.Method == http.MethodPost {
				seriesH.UploadImage(w, r)
			} else {
				methodNotAllowed(w)
			}
			return
		}

		if strings.Contains(path, "/rating") {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) == 4 && r.Method == http.MethodDelete {
				ratingsH.Delete(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				ratingsH.Get(w, r)
			case http.MethodPost:
				ratingsH.Create(w, r)
			default:
				methodNotAllowed(w)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			seriesH.GetByID(w, r)
		case http.MethodPut:
			seriesH.Update(w, r)
		case http.MethodDelete:
			seriesH.Delete(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	handler := middleware.CORS(mux)

	port := getEnv("PORT", "8080")
	log.Printf("Server running on http://localhost:%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, `{"error":"Method not allowed"}`)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

var _ *sql.DB
