package service

import (
	"IM/models"
	"IM/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

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
type ChatRedisMsg struct {
	SendUserID uint   `json:"send_user_id"`
	RecvUserID uint   `json:"recv_user_id"`
	GroupID    uint   `json:"group_id"`
	Content    string `json:"content"`
	MsgType    int    `json:"msg_type"` // 1=文本消息，2=图片消息，3=文件消息
	Timestamp  int64  `json:"timestamp"`
}
type ChatHistoryQuery struct {
	TargetID uint `form:"target_id" binding:"required,min=1" comment:"目标用户ID"`
	Page     int  `form:"page" binding:"required,min=1" comment:"页码"`
	Size     int  `form:"size" binding:"required,min=1,max=100" comment:"每页数量"`
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

	// 检查用户名是否已存在
	existing, _ := models.FindUserByName(user.Name)
	if existing != nil && existing.ID != 0 {
		respFail(c, 400, "用户名已存在")
		return
	}

	// 自动生成唯一身份标识
	user.Identity = utils.Md5Encode(fmt.Sprintf("%d%s", time.Now().UnixNano(), user.Name))
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


// WS升级器
var upGrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 在线客户端管理：按用户ID索引，支持多设备登录
var (
	userConns = make(map[uint]map[*websocket.Conn]struct{})
	connMu    sync.RWMutex
	writeMu   sync.Mutex // 并发写ws锁，防止1006异常关闭
)

// sendToUser 向指定用户的所有在线设备推送消息
func sendToUser(recvUserID uint, msg []byte) {
	connMu.RLock()
	conns, ok := userConns[recvUserID]
	connMu.RUnlock()
	if !ok {
		return
	}
	for conn := range conns {
		writeMu.Lock()
		err := conn.WriteMessage(websocket.TextMessage, msg)
		writeMu.Unlock()
		if err != nil {
			log.Printf("推送消息给用户%d失败: %v", recvUserID, err)
			connMu.Lock()
			delete(conns, conn)
			if len(userConns[recvUserID]) == 0 {
				delete(userConns, recvUserID)
			}
			connMu.Unlock()
			_ = conn.Close()
		}
	}
}

// Broadcast 接收redis消息，路由给目标用户 + 异步入库
func Broadcast(payload string) {
	var msg ChatRedisMsg
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Println("解析Redis消息JSON失败：", err)
		return
	}

	// 异步入库
	go func() {
		chatMsg := &models.ChatMsg{
			SendUserID: msg.SendUserID,
			RecvUserID: msg.RecvUserID,
			GroupID:    msg.GroupID,
			Content:    msg.Content,
			MsgType:    msg.MsgType,
		}
		if err := models.CreateChatMsg(chatMsg); err != nil {
			log.Println("models保存聊天消息失败：", err)
		}
	}()

	// 推送给接收方 + 回显给发送方（确认送达）
	msgBytes, _ := json.Marshal(msg)
	sendToUser(msg.RecvUserID, msgBytes)
	sendToUser(msg.SendUserID, msgBytes)
}

// SendMsg WebSocket 连接处理
func SendMsg(c *gin.Context) {
	// ========== Token鉴权 ==========
	tokenStr := c.Query("token")
	if tokenStr == "" {
		log.Println("WS拒绝连接：未携带token参数")
		return
	}
	claims, err := utils.ParseToken(tokenStr)
	if err != nil {
		log.Println("WS拒绝连接：token非法或过期, err:", err)
		return
	}
	loginUserID := claims.UserID
	log.Printf("用户ID:%d 建立WS连接", loginUserID)

	// 1. 协议升级
	conn, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级协议错误：", err)
		return
	}

	// 2. 上线：加入在线集合
	connMu.Lock()
	if userConns[loginUserID] == nil {
		userConns[loginUserID] = make(map[*websocket.Conn]struct{})
	}
	userConns[loginUserID][conn] = struct{}{}
	connMu.Unlock()

	// 3. 下线清理
	defer func() {
		connMu.Lock()
		delete(userConns[loginUserID], conn)
		if len(userConns[loginUserID]) == 0 {
			delete(userConns, loginUserID)
		}
		connMu.Unlock()
		_ = conn.Close()
		log.Printf("用户ID:%d 离线", loginUserID)
	}()

	// 心跳保活
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 循环读取客户端消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息错误：", err)
			break
		}

		// 解析客户端JSON消息
		var clientMsg ChatRedisMsg
		if err := json.Unmarshal(message, &clientMsg); err != nil || clientMsg.RecvUserID == 0 {
			log.Printf("用户%d消息格式错误", loginUserID)
			writeMu.Lock()
			_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"请使用JSON格式: {\"recv_user_id\":123,\"content\":\"hello\"}"}`))
			writeMu.Unlock()
			continue
		}

		clientMsg.SendUserID = loginUserID
		if clientMsg.MsgType == 0 {
			clientMsg.MsgType = 1 // 默认文本消息
		}
		clientMsg.Timestamp = time.Now().UnixMilli()

		// 发布到Redis，由Broadcast处理持久化和推送
		jsonBuf, err := json.Marshal(clientMsg)
		if err != nil {
			log.Println("消息序列化失败：", err)
			continue
		}
		_ = utils.Publish(c.Request.Context(), "websocket", string(jsonBuf))
	}
}
