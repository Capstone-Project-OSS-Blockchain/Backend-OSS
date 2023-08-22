package upload

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	_ "github.com/minio/minio-go/v7/pkg/credentials"

	minioconnections "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/connections"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	minioClient, err := minioconnections.InitMinIOClient()
	if err != nil {
		http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
		return
	}

	// Generate a UUID for the filename
	filename := uuid.New().String() + ".pdf"

	// Create a new PDF file
	file, err := os.Create(filename)
	if err != nil {
		http.Error(w, "Failed to create PDF file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Here you would add code to generate the PDF content using a library like gopdf or gofpdf.
	// For demonstration purposes, let's assume we are writing some basic content to the PDF.
	_, err = file.WriteString("This is a generated PDF content.")
	if err != nil {
		http.Error(w, "Failed to write content to PDF file", http.StatusInternalServerError)
		return
	}

	// Upload the generated PDF file to MinIO
	bucketName := "services-bucket"
	objectName := filename

	file.Seek(0, 0) // Reset file pointer to the beginning before uploading

	_, err = minioClient.PutObject(r.Context(), bucketName, objectName, file, -1, minio.PutObjectOptions{
		ContentType: "application/pdf",
	})
	if err != nil {
		http.Error(w, "Error uploading PDF file to MinIO", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "PDF file generated and uploaded successfully!")
}
