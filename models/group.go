package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupBasic struct {
	gorm.Model
	Name       string    `gorm:"comment:群名称"`
	OwnerID    uint      `gorm:"comment:群主ID"`
	CreateTime time.Time `gorm:"comment:创建时间"`
	Topic      string    `gorm:"comment:群头像"`
}

func (table *GroupBasic) TableName() string { return "group_basic" }
