package connections

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinIOClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT") 
	accessKey := os.Getenv("MINIO_ACCESSKEY")
	secretKey := os.Getenv("MINIO_SECRETKEY")
	useSSL := true

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}
