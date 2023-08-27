package ownership

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/gorm"
)

var (
	mysqlDB   *gorm.DB
	secretKey = os.Getenv("SECRET_KEY")
	validate  = validator.New()
)

func InitDB(db *gorm.DB) {
	mysqlDB = db
}

func GetOwnershipByUserID(w http.ResponseWriter, r *http.Request) {
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
	var ownerships []models.Ownership

	if err := mysqlDB.Table("owners").Where("id = ?", userID).Find(&ownerships).Error; err != nil {
		return
	}

	var files []map[string]interface{}
	for _, ownership := range ownerships {
		fileData := map[string]interface{}{
			"filename":  ownership.Filename,
			"timestamp": ownership.Timestamp,
		}
		files = append(files, fileData)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":  userID,
		"files":   files,
		"message": "The PDF File obtained!",
	})
}
