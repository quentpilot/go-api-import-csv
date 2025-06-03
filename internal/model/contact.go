package model

import "gorm.io/gorm"

type Contact struct {
	gorm.Model
	ReqId     string `gorm:"index:idx_req_id"`
	Phone     string
	Firstname string
	Lastname  string
}
