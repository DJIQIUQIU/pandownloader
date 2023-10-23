package utils

import (
	"crypto/md5"
	"fmt"
	"git-sec.com/pandownloader/logger"
	"github.com/Unknwon/goconfig"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"strings"
)

var Cfg *goconfig.ConfigFile

/**
 * 获取配置文件
 */
func GetConfigIni(filepath string) (err error) {
	config, err := goconfig.LoadConfigFile(filepath)
	if err != nil {
		fmt.Println("配置文件读取错误,找不到配置文件", err)
		return err
	}
	Cfg = config
	return nil
}

/**
 * session认证拦截
 */
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := url.Parse(c.Request.RequestURI)
		if err != nil {
			logger.GetLogger().Println("err", err)
		}
		logger.GetLogger().Println("u.path", u.Path)
		if strings.Contains(u.Path, "auth") || strings.Contains(u.Path, "share") {
			c.Next()
			return
		}
		//  验证参数
		token := ""
		if strings.Contains(u.Path, "download") {
			token = c.Query("token")
		} else {
			token = c.Request.Header.Get("P-Token")
		}
		//fmt.Println("auth token", token)
		//if token == "" {
		//	c.JSON(200, gin.H{
		//		"code": 1000,
		//		"msg":  "token must be not null",
		//	})
		//	return
		//}
		// 查找session是否存在
		session := sessions.Default(c)
		logger.GetLogger().Println("token_pan_", token)
		v := session.Get("token_pan_" + token)
        logger.GetLogger().Println("v:", v)
		if v == nil {
			// c.Abort()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "2000",
				"message": "not authed",
			})
			return
		}
		c.Next()
	}
}

func GetCorsConfig() cors.Config {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://pan-download.djicorp.com", "http://localhost:55581", "http://localhost", "http://10.60.18.220:8080"}
	config.AllowMethods = []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"}
	config.AllowCredentials = true
	config.AllowHeaders = []string{"x-requested-with", "Content-Type", "AccessToken", "X-CSRF-Token", "X-Token", "Authorization", "token"}
	return config
}

func Md5V2(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str // 判断所给路径文件/文件夹是否存在
}
