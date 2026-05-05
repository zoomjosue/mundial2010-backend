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
	// Connect to DB
	database, err := db.Connect()
	if err != nil {
		log.Fatal("Could not connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.Migrate(database); err != nil {
		log.Fatal("Migration failed:", err)
	}

	// Init handlers
	seriesH := handlers.NewSeriesHandler(database)
	ratingsH := handlers.NewRatingsHandler(database)

	// Router
	mux := http.NewServeMux()

	// Serve uploads statically
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// Serve Swagger UI and spec
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/swagger-ui.html", http.StatusMovedPermanently)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","message":"Mundial 2010 API running"}`)
	})

	// Series routes
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

		// POST /series/:id/image
		if strings.HasSuffix(path, "/image") {
			if r.Method == http.MethodPost {
				seriesH.UploadImage(w, r)
			} else {
				methodNotAllowed(w)
			}
			return
		}

		// /series/:id/rating and /series/:id/rating/:ratingId
		if strings.Contains(path, "/rating") {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			// /series/:id/rating/:ratingId
			if len(parts) == 4 && r.Method == http.MethodDelete {
				ratingsH.Delete(w, r)
				return
			}
			// /series/:id/rating
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

		// /series/:id
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

	// Apply middleware
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

// Needed for db package's sql.Open with postgres driver
var _ *sql.DB
