package model

import "gorm.io/gorm"

type Contact struct {
	gorm.Model
	Phone     string
	Firstname string
	Lastname  string
}
