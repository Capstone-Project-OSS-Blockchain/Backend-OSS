package download

import (
	"context"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	_ "github.com/minio/minio-go/v7/pkg/credentials"

	minioconnections "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/connections"
	"github.com/gorilla/mux"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := mux.Vars(r)["filename"]

	minioClient, err := minioconnections.InitMinIOClient()
	if err != nil {
		http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
		return
	}
	
	//nama bucket untuk minio
	bucket := "wasabi-bucket"

	object, err := minioClient.GetObject(context.Background(), bucket, fileName, minio.GetObjectOptions{})
																	//bucket name diubah menggunakan bucket local, tapi tetap masih fail
	if err != nil {
		http.Error(w, "Failed to retrieve file from MinIO", http.StatusInternalServerError)
		return
	}
	defer object.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)

	// Retrieve content type from the response
	info, err := object.Stat()
	if err != nil {
		http.Error(w, "Failed to get object info", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", info.ContentType)

	// Copy the MinIO object's content to the response writer
	_, err = io.Copy(w, object)
	if err != nil {
		http.Error(w, "Failed to stream file content", http.StatusInternalServerError)
		return
	}
}
