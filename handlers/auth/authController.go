package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/Capstone-Project-OSS-Blockchain/Backend-OSS/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var (
	mysqlDB   *gorm.DB
	secretKey = os.Getenv("SECRET_KEY")
	validate  = validator.New()
)

func InitDB(db *gorm.DB) {
	mysqlDB = db
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid request",
		})
		return
	}

	var validationErrors []string

	if err := validate.Struct(newUser); err != nil {
		for _, validationErr := range err.(validator.ValidationErrors) {
			fieldName := validationErr.Field()
			var errorMessage string

			switch validationErr.Tag() {
			case "required":
				errorMessage = fmt.Sprintf("%s is required", fieldName)
			case "email":
				errorMessage = fmt.Sprintf("%s is not a valid email address", fieldName)
			case "min":
				errorMessage = fmt.Sprintf("%s must be at least %s characters long", fieldName, validationErr.Param())
			case "max":
				errorMessage = fmt.Sprintf("%s must be at most %s characters long", fieldName, validationErr.Param())
			default:
				errorMessage = fmt.Sprintf("Invalid %s", fieldName)
			}

			validationErrors = append(validationErrors, errorMessage)
		}
	}

	if !IsStrongPassword(newUser.Password) {
		validationErrors = append(validationErrors, "Password must contain at least 1 special character")
	}

	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid input data",
			"errors":  validationErrors,
		})
		return
	}

	var cekUser models.User
	if err := mysqlDB.Where("email = ?", newUser.Email).First(&cekUser).Error; err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Email already exists",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("Error querying user in MySQL:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Failed to register user",
		})
		return
	}

	newUser.Id = uuid.New().String()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Failed to register user",
		})
		return
	}

	newUser.Password = string(hashedPassword)

	if err := mysqlDB.Create(&newUser).Error; err != nil {
		fmt.Println("Error creating user in MySQL:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Failed to register user",
		})
		return
	}

	response := map[string]interface{}{
		// "data":    newUser,
		"message": "User registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid request",
		})
		return
	}

	var user models.User

	if err := mysqlDB.Where("email = ?", data["email"]).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "User not found",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Error retrieving user data",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data["password"])); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Incorrect password",
		})
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Id,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(secretKey))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Couldn't login",
		})
		return
	}

	// w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":   token,
		"message": "Login Success",
	})
}

func IsStrongPassword(password string) bool {
	const (
		minSpecialChars = 1
	)

	specialChars := 0
	for _, char := range password {
		if unicode.IsPunct(char) || unicode.IsSymbol(char) {
			specialChars++
		}
	}

	return specialChars >= minSpecialChars
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Authorization token missing",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Invalid token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Invalid token",
			})
			return
		}

		Id, ok := claims["iss"].(string)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Invalid token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), "id", Id)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	Id, ok := r.Context().Value("id").(string)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Unauthorized",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Protected route accessed",
		"id":      Id,
	})
}
