package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	_ "os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/jung-kurt/gofpdf"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	_ "github.com/minio/minio-go/v7/pkg/credentials"

	minioconnections "github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/connections"
	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/models"
)

var (
	mysqlDB   *gorm.DB
	secretKey = os.Getenv("SECRET_KEY")
)

func InitDB(db *gorm.DB) {
	mysqlDB = db
}

func generateNIB() string {
	// Generate a random NIB number with a length of 7
	nib := ""
	for i := 0; i < 13; i++ {
		nib += strconv.Itoa(rand.Intn(10)) // Requires "math/rand" import
	}
	return nib
}

func generatePDF() (*bytes.Buffer, string, error) {

	// Generate a UUID for the filename
	filename := uuid.New().String() + ".pdf"

	// Create a buffer to store PDF content
	pdfBuffer := new(bytes.Buffer)

	// Initialize PDF generation using gofpdf
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add PNG image to the middle of the header
	imgURL := "https://diskop.bandungkab.go.id/themes/diskop/frontend/images/oss.png"
	imgResp, err1 := http.Get(imgURL)
	if err1 != nil {
		return nil, "", fmt.Errorf("error downloading image: %v", err1)
	}
	defer imgResp.Body.Close()

	imgData, err2 := io.ReadAll(imgResp.Body)
	if err2 != nil {
		return nil, "", fmt.Errorf("error reading image: %v", err2)
	}

	pdf.RegisterImageOptionsReader("image", gofpdf.ImageOptions{ImageType: "png"}, bytes.NewReader(imgData))
	pdf.ImageOptions("image", 50, 10, 100, 0, false, gofpdf.ImageOptions{}, 0, "")

	// Add text under the image
	pdf.SetFont("Arial", "B", 16)
	pdf.SetY(120)                          // Set Y position for text under the image
	pdf.Cell(0, 10, "NIB: "+generateNIB()) // Generate and add NIB

	// Add additional text under the NIB with left alignment
	pdf.SetFont("Arial", "", 12) // Set font for additional text
	pdf.SetX(10)                 // Set X position for left alignment
	pdf.SetY(pdf.GetY() + 10)    // Move down a bit
	additionalText := "This is A Generated PDF Document"
	pdf.Cell(0, 10, additionalText)

	// Output PDF content

	pdf.Output(pdfBuffer)

	return pdfBuffer, filename, nil
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	minioClient, err := minioconnections.InitMinIOClient()
	if err != nil {
		http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
		return
	}

	pdfBuffer, filename, err := generatePDF()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bucketName := "services-bucket"
	objectName := filename

	_, err = minioClient.PutObject(r.Context(), bucketName, objectName, pdfBuffer, int64(pdfBuffer.Len()), minio.PutObjectOptions{
		ContentType: "application/pdf",
	})
	if err != nil {
		http.Error(w, "Error uploading PDF file to MinIO", http.StatusInternalServerError)
		return
	}

	// fmt.Fprintln(w, filename+" PDF file generated and uploaded successfully!")

	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		http.Error(w, "Failed to parse token", http.StatusBadRequest)
		return
	}

	if !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Failed to extract claims from token", http.StatusBadRequest)
		return
	}

	userID, ok := claims["iss"].(string)
	if !ok {
		http.Error(w, "Invalid user_id claim", http.StatusBadRequest)
		return
	}
	// currentTime := time.Now()
	// date := currentTime.Format("15:04:05 02-01-2006")
	date := time.Now()
	newOwnership := models.Ownership{
		UserId:    userID,
		Filename:  filename,
		Timestamp: date,
	}

	if err := mysqlDB.Table("owners").Create(&newOwnership).Error; err != nil {
		fmt.Println("Error creating ownership record:", err)
		http.Error(w, "Failed to create ownership record", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"FileName": filename,
		"message":  "The PDF File is generated and Uploaded Successfully!",
	})

	// Close the buffer
	defer pdfBuffer.Reset()
}
