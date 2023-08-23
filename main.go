package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/connections"
	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/handlers/auth"
	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/handlers/download"
	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/handlers/upload"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	connections.Connect()

	r := mux.NewRouter()

	auth.InitDB(connections.DB)
	r.HandleFunc("/register", auth.RegisterUser).Methods("POST")
	r.HandleFunc("/login", auth.Login).Methods("POST")
	r.HandleFunc("/upload", upload.UploadHandler).Methods("POST")
	//r.HandleFunc("/download/{filename}", download.DownloadHandler).Methods("GET")
	//r.HandleFunc("/protected", auth.ProtectedRoute).Methods("GET").Handler(auth.AuthMiddleware(http.HandlerFunc(auth.ProtectedRoute)))

	// r.HandleFunc("/upload", upload.UploadHandler).Methods("POST").Handler(auth.AuthMiddleware(http.HandlerFunc(upload.UploadHandler)))
	r.HandleFunc("/download/{filename}", download.DownloadHandler).Methods("GET").Handler(auth.AuthMiddleware(http.HandlerFunc(download.DownloadHandler)))

	http.Handle("/", r)

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
