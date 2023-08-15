package main

import (
	
	"fmt"

	"net/http"

	"github.com/gorilla/mux"
	upload "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/handlers/upload"
	download "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/handlers/download"
	
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/upload", upload.UploadHandler).Methods("POST")
	router.HandleFunc("/download/{filename}", download.DownloadHandler).Methods("GET")
	// http.Handle("/", router)

	err := http.ListenAndServe(":4000", router)
	fmt.Println("listen on localhost :4000")

	if err != nil {
		fmt.Println(err)
	}
}