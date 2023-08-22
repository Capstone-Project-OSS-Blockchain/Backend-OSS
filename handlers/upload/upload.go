package upload

import (
    "bytes"
    "fmt"
    "net/http"
    _ "os"

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

    // Create a buffer to store PDF content
    pdfBuffer := new(bytes.Buffer)

    // Here you would add code to generate the PDF content using a library like gopdf or gofpdf.
    // For demonstration purposes, let's assume we are writing some basic content to the PDF.
    pdfBuffer.WriteString("This is a generated PDF content.")

    // Upload the generated PDF buffer to MinIO
    bucketName := "wasabi-bucket"
    objectName := filename

    _, err = minioClient.PutObject(r.Context(), bucketName, objectName, pdfBuffer, int64(pdfBuffer.Len()), minio.PutObjectOptions{
        ContentType: "application/pdf",
    })
    if err != nil {
        http.Error(w, "Error uploading PDF file to MinIO", http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "PDF file generated and uploaded successfully!")

    // Close the buffer
    defer pdfBuffer.Reset()
}