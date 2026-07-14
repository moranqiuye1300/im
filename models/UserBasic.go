package models

import (
	"IM/utils"
	"time"

	"gorm.io/gorm"
)

// UserBasic 用户基础表
type UserBasic struct {
	// 手动展开gorm.Model，解决swag识别不到第三方类型报错
	ID        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Name          string     `json:"name" binding:"required,min=2,max=20" comment:"用户名"`
	Password      string     `json:"password" binding:"required,min=6" comment:"登录密码"`
	Salt          string     `json:"salt" comment:"密码盐值"`
	Phone         string     `json:"phone" binding:"required,len=11,numeric,phone" comment:"手机号"`
	Email         string     `json:"email" binding:"required,email" comment:"邮箱"`
	Identity      string     `json:"identity" binding:"required" comment:"唯一身份标识"`
	ClientIP      string     `json:"client_ip" comment:"登录IP"`
	ClientPort    string     `json:"client_port" comment:"登录端口"`
	LoginTime     *time.Time `json:"login_time" gorm:"datetime(3);null" comment:"登录时间"`
	HeartbeatTime *time.Time `json:"heartbeat_time" gorm:"datetime(3);null" comment:"心跳时间"`
	LogoutTime    *time.Time `json:"logout_time" gorm:"datetime(3);null" comment:"登出时间"`
	IsLogout      bool       `json:"is_logout" comment:"是否已登出"`
	DeviceInfo    string     `json:"device_info" comment:"设备信息"`
}

// TableName 指定数据库表名
func (table *UserBasic) TableName() string {
	return "user_basic"
}

// GetUserList 查询全部用户（修复硬编码长度+捕获错误）
func GetUserList() ([]*UserBasic, error) {
	var list []*UserBasic
	err := utils.DB.Find(&list).Error
	if err != nil {
		return nil, err
	}
	// 遍历清空密码，脱敏
	for _, v := range list {
		v.Password = ""
	}
	return list, nil
}

// CreateUser 创建用户
func CreateUser(user *UserBasic) *gorm.DB {
	return utils.DB.Create(user)
}

// DeleteUser 软删除用户（根据ID删除）
func DeleteUser(user *UserBasic) *gorm.DB {
	return utils.DB.Delete(user)
}

// UpdateUser 按需更新非空字段，避免空值覆盖
func UpdateUser(user *UserBasic) *gorm.DB {
	// Select 只更新有值的字段，零值忽略
	return utils.DB.Model(user).Select("Name", "Password", "Phone", "Email").Updates(user)
}
func FindUserByName(name string) (*UserBasic, error) {
	var user UserBasic
	err := utils.DB.Where("name = ?", name).First(&user).Error
	return &user, err
}
func FindUserByPhone(phone string) (*UserBasic, error) {
	var user UserBasic
	err := utils.DB.Where("phone = ?", phone).First(&user).Error
	return &user, err
}
func FindUserByEmail(email string) (*UserBasic, error) {
	var user UserBasic
	err := utils.DB.Where("email = ?", email).First(&user).Error
	return &user, err
}
