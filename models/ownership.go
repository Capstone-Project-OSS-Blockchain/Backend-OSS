package models

type Ownership struct {
	Id        uint   `json:"id" gorm:"primary_key;column:ownerID;auto_increment"`
	UserId    string `json:"userId" gorm:"column:id"`
	Filename  string `json:"filename"`
	Timestamp string `json:"date" gorm:"column:timestamp`
}
