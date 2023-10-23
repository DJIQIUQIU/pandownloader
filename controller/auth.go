package controller

import (
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"

	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/model"
)

// 请求参数
type BodyJson struct {
	Password string `json:"password"`
}

// 验证密码, 兼容2i, 2id两个版本
func compareHash(version int, password string, hash string) (match bool, err error) {
	var otherKey []byte
	params, salt, key, err := argon2id.DecodeHash(hash)
	switch version {
	case 2:
		otherKey = argon2.Key([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
	default:
		otherKey = argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
	}
	keyLen := int32(len(key))
	otherKeyLen := int32(len(otherKey))

	if subtle.ConstantTimeEq(keyLen, otherKeyLen) == 0 {
		return false, nil
	}
	if subtle.ConstantTimeCompare(key, otherKey) == 1 {
		return true, nil
	}
	return false, nil
}

/**
 * 验证密码
 */
func Auth(c *gin.Context) {
	var (
		json BodyJson
	)
    
	// 参数校验
	token := c.Request.Header.Get("P-Token")
	fmt.Println("auth token", token)
	if token == "" {
		c.JSON(200, gin.H{
			"code": 1000,
			"msg":  "token must be not null",
		})
		return
	}
	err := c.BindJSON(&json)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1000,
			"msg":  "parameter invalid",
		})
		return
	}
	fmt.Println("password:", json.Password, "token:", token)
	// 查询密码
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1001,
			"msg":  "token not found:",
		})
		return
	}
	ok := true
	logger.GetLogger().Println("hash:", share.Permissions, ':', share.Password)
	if share.Password != "" {
		// 密码以|分割，前面为next cloud的版本
		hashSplit := strings.Split(share.Password, "|")
		hash := hashSplit[1]
		// 密码检查
		// ok, err := argon2id.ComparePasswordAndHash(json.Password, hash)
		version, _ := strconv.Atoi(hashSplit[0])
		ok, err = compareHash(version, json.Password, hash)
	}

	if ok {
		// 以token保存
		session := sessions.Default(c)
		logger.GetLogger().Println("ok:", ok)
		session.Set("token_pan_"+token, token)
		err = session.Save()
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
		})
		logger.GetLogger().Println("s:", token, "err", err)
	} else {
		logger.GetLogger().Println("err:", err)
		c.JSON(200, gin.H{
			"code": 1002,
			"msg":  "password do not match",
		})
	}
}

/**
 * token分享权限
 */
func Share(c *gin.Context) {
	token := c.Request.Header.Get("P-Token")
	if token == "" {
		c.JSON(200, gin.H{
			"code": 1000,
			"msg":  "Share: parameter invalid",
		})
		return
	}
	fmt.Println("token:", token)
	// 查询share数据
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1001,
			"msg":  "token not found:",
		})
		return
	} else {
		needAuth := true
		if share.Password == "" {
			needAuth = false
		}
		data := gin.H{
			"permissions": share.Permissions,
			"target":      strings.Replace(share.FileTarget, "/", "", 1),
			"share_auth":  needAuth,
		}
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
			"data": data,
		})
	}
}
