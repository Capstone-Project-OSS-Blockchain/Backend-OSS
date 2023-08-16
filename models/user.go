package models

type User struct {
	Id       string `json:"id" gorm:"primary_key;column:id"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username"`
	Password string `json:"password" validate:"required,min=8,max=16"`
}
