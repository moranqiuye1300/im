package service

import (
	"IM/models"
	"IM/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MsgResp 统一返回结构体，给Swag识别字段
type MsgResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// DeleteUserReq 删除用户入参
type DeleteUserReq struct {
	ID uint `json:"id" binding:"required,min=1" comment:"用户ID"`
}

// UpdateUserReq 更新用户入参
type UpdateUserReq struct {
	ID       uint   `json:"id" binding:"required,min=1" comment:"用户ID"`
	Name     string `json:"name" binding:"omitempty,min=2,max=20" comment:"用户名"`
	Password string `json:"password" binding:"omitempty,min=6" comment:"登录密码"`
	Phone    string `json:"phone" binding:"omitempty,len=11,numeric" comment:"手机号"`
	Email    string `json:"email" binding:"omitempty,email" comment:"邮箱"`
}

// NameQuery 按名称查询GET参数
type NameQuery struct {
	Name string `form:"name" binding:"required,min=2,max=20" comment:"用户名"`
}

// PhoneQuery 按手机号查询GET参数
type PhoneQuery struct {
	Phone string `form:"phone" binding:"required,len=11,numeric" comment:"手机号"`
}

// EmailQuery 按邮箱查询GET参数
type EmailQuery struct {
	Email string `form:"email" binding:"required,email" comment:"邮箱"`
}

// ---------------- 统一封装响应工具函数 ----------------
func respOk(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, MsgResp{
		Code:    200,
		Message: msg,
		Data:    data,
	})
}
func respFail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, MsgResp{
		Code:    code,
		Message: msg,
	})
}
func respServerErr(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, MsgResp{
		Code:    500,
		Message: msg,
	})
}

// GetUserList 获取用户列表
// @Summary 查询全部用户
// @Description 获取系统所有用户基础信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Success 200 {object} MsgResp
// @Router /user/list [get]
func GetUserList(c *gin.Context) {
	userList, err := models.GetUserList()
	if err != nil {
		respServerErr(c, "查询用户失败："+err.Error())
		return
	}
	respOk(c, "查询成功", userList)
}

// CreateUser 创建用户
// @Summary 新增用户
// @Description 传入用户JSON信息创建账号，密码自动加密
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param user body models.UserBasic true "用户信息表单"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp "JSON参数解析失败"
// @Failure 500 {object} MsgResp "数据库写入失败"
// @Router /user/create [post]
func CreateUser(c *gin.Context) {
	var user models.UserBasic
	if err := c.ShouldBindJSON(&user); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}

	// 修复：加密逻辑放在入库之前
	user.Salt = utils.RandomSalt(16) // 生成随机盐
	user.Password = utils.MakePassword(user.Password, user.Salt)

	result := models.CreateUser(&user)
	if result.Error != nil {
		respServerErr(c, "用户创建失败："+result.Error.Error())
		return
	}
	respOk(c, "用户创建成功", nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID软删除用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param body body DeleteUserReq true "仅需用户id"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/delete [post]
func DeleteUser(c *gin.Context) {
	var req DeleteUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}
	user := models.UserBasic{ID: req.ID}
	result := models.DeleteUser(&user)
	if result.Error != nil {
		respServerErr(c, "用户删除失败："+result.Error.Error())
		return
	}
	respOk(c, "用户删除成功", nil)
}

// UpdateUser 更新用户
// @Summary 更新用户信息
// @Description 根据用户ID更新指定字段
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param body body UpdateUserReq true "用户更新参数"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/update [post]
func UpdateUser(c *gin.Context) {
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}
	user := models.UserBasic{
		ID:       req.ID,
		Name:     req.Name,
		Password: req.Password,
		Phone:    req.Phone,
		Email:    req.Email,
	}
	if req.Password != "" {
		user.Salt = utils.RandomSalt(16)
		user.Password = utils.MakePassword(req.Password, user.Salt)
	}
	result := models.UpdateUser(&user)
	if result.Error != nil {
		respServerErr(c, "用户更新失败："+result.Error.Error())
		return
	}
	respOk(c, "用户更新成功", nil)
}

// FindUserByName 根据用户名查询用户
// @Summary 按用户名查询
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param name query string true "用户名"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/find/name [get]
func FindUserByName(c *gin.Context) {
	var req NameQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}
	user, err := models.FindUserByName(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respOk(c, "未查询到该用户", nil)
			return
		}
		respServerErr(c, "查询用户失败："+err.Error())
		return
	}
	respOk(c, "查询成功", user)
}

// FindUserByPhone 根据手机号查询用户
// @Summary 按手机号查询
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param phone query string true "手机号"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/find/phone [get]
func FindUserByPhone(c *gin.Context) {
	var req PhoneQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}
	user, err := models.FindUserByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respOk(c, "未查询到该用户", nil)
			return
		}
		respServerErr(c, "查询用户失败："+err.Error())
		return
	}
	respOk(c, "查询成功", user)
}

// FindUserByEmail 根据邮箱查询用户
// @Summary 按邮箱查询
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param email query string true "邮箱"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/find/email [get]
func FindUserByEmail(c *gin.Context) {
	var req EmailQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}
	user, err := models.FindUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respOk(c, "未查询到该用户", nil)
			return
		}
		respServerErr(c, "查询用户失败："+err.Error())
		return
	}
	respOk(c, "查询成功", user)
}

// LoginReq 登录请求参数
type LoginReq struct {
	Name     string `json:"name" binding:"required,min=2,max=20" comment:"用户名"`
	Password string `json:"password" binding:"required,min=6" comment:"登录密码"`
}

// LoginResp 登录响应
type LoginResp struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户名+密码登录，返回JWT token
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录参数"
// @Success 200 {object} MsgResp
// @Failure 400 {object} MsgResp
// @Failure 500 {object} MsgResp
// @Router /user/login [post]
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respFail(c, 400, "参数校验失败："+err.Error())
		return
	}

	// 查找用户
	user, err := models.FindUserByName(req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respFail(c, 401, "用户名或密码错误")
			return
		}
		respServerErr(c, "登录失败："+err.Error())
		return
	}

	// 验证密码
	if !utils.ValidatePassword(req.Password, user.Salt, user.Password) {
		respFail(c, 401, "用户名或密码错误")
		return
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Name)
	if err != nil {
		respServerErr(c, "token生成失败："+err.Error())
		return
	}

	respOk(c, "登录成功", LoginResp{
		Token:    token,
		UserID:   user.ID,
		UserName: user.Name,
	})
}
