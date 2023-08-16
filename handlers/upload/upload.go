package upload

import (
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
	_ "github.com/minio/minio-go/v7/pkg/credentials"

	minioconnections "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/connections"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	err := r.ParseMultipartForm(1024)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	minioClient, err := minioconnections.InitMinIOClient()
	if err != nil {
		http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
		return
	}

	// Loop through uploaded files
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Failed to open uploaded file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Upload the file to MinIO
			bucketName := "dzaky-upload"
			//bucket name diubah menggunakan bucket local, tapi tetap masih fail
			objectName := fileHeader.Filename

			_, err = minioClient.PutObject(r.Context(), bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
				ContentType: fileHeader.Header.Get("Content-Type"),
			})
			if err != nil {
				http.Error(w, "Error uploading file to MinIO", http.StatusInternalServerError)
				return
			}
		}
	}

	fmt.Fprintln(w, "Files uploaded successfully!")
}
