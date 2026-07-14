package models

import (
	"IM/utils"
	"log"

	"gorm.io/gorm"
)

type ChatMsg struct {
	gorm.Model
	SendUserID uint   `gorm:"comment:发送用户ID"`
	RecvUserID uint   `gorm:"comment:接收用户ID，私聊；群聊存0"`
	GroupID    uint   `gorm:"comment:群ID，私聊存0"`
	Content    string `gorm:"type:text;comment:消息内容"`
	MsgType    int    `gorm:"comment:1文本 2图片 3文件"`
}

func (table *ChatMsg) TableName() string {
	return "chat_msg"
}

// CreateChatMsg 持久化聊天消息，仅数据库操作
func CreateChatMsg(msg *ChatMsg) error {
	res := utils.DB.Create(msg)
	if res.Error != nil {
		log.Printf("保存聊天记录失败：%v", res.Error)
		return res.Error
	}
	return nil
}
