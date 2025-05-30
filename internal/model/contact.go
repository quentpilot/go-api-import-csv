package model

import "gorm.io/gorm"

type Contact struct {
	gorm.Model
	ReqId     string
	Phone     string
	Firstname string
	Lastname  string
}
