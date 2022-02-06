package controller

import (
	"net/http"

	"example.com/m/ws"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func EnableCookieSession() gin.HandlerFunc {
	store := cookie.NewStore([]byte("secret"))
	return sessions.Sessions("test", store)
}

func Index(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username != nil {
		rooms := []map[string]interface{}{
			{"id": 1, "num": ws.GetOnlineRoomUserCount(1)},
			{"id": 2, "num": ws.GetOnlineRoomUserCount(2)},
		}
		c.HTML(http.StatusOK, "home.html", gin.H{
			"username": username,
			"rooms":    rooms,
		})
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func Home(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}
	rooms := []map[string]interface{}{
		{"id": 1, "num": ws.GetOnlineRoomUserCount(1)},
		{"id": 2, "num": ws.GetOnlineRoomUserCount(2)},
	}
	c.HTML(http.StatusOK, "home.html", gin.H{
		"rooms":    rooms,
		"username": username,
	})
}

func Room(c *gin.Context) {
	roomId := c.Param("room_id")
	rooms := []string{"1", "2"}
	if !InArray(roomId, rooms) {
		c.Redirect(http.StatusFound, "/room/1")
		return
	}
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}
	c.HTML(http.StatusOK, "room.html", gin.H{
		"roomId":   roomId,
		"username": username,
	})
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	if username == "" {
		c.JSON(http.StatusOK, gin.H{"code": 5000, "msg": "用户名不能为空"})
		return
	}
	session := sessions.Default(c)
	session.Set("username", username)
	session.Save()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}

func InArray(needle interface{}, hystack interface{}) bool {
	switch key := needle.(type) {
	case string:
		for _, item := range hystack.([]string) {
			if key == item {
				return true
			}
		}
	case int:
		for _, item := range hystack.([]int) {
			if key == item {
				return true
			}
		}
	case int64:
		for _, item := range hystack.([]int64) {
			if key == item {
				return true
			}
		}
	default:
		return false
	}
	return false
}
