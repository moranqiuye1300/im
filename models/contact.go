package models

import "gorm.io/gorm"

type Contact struct {
	gorm.Model
	OwnerID  uint `gorm:"comment:拥有者ID"`
	TargetID uint `gorm:"comment:目标ID"`
	Type     int  `gorm:"comment:1好友 2群组"`
	Status   int  `gorm:"comment:1申请中 2同意 3拒绝"`
}

func (table *Contact) TableName() string { return "contact" }
