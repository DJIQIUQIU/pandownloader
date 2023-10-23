package main

import (
	"fmt"

	// "github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	ginR "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"

    "git-sec.com/pandownloader/controller"
	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/module/cache"
    "git-sec.com/pandownloader/module/utils"
	pubsub "git-sec.com/pandownloader/module/db" // 初始化mysql， redis
)

func main() {
	logger.GetLogger().Print("run")
	r := gin.Default()
	store, _ := ginR.NewStoreWithPool(cache.RedisClient, []byte("secret"))
	fmt.Println("store", store)
	r.Use(sessions.Sessions("pansession", store))
	// r.Use(cors.New(utils.GetCorsConfig())) //跨域

    go pubsub.Subscribe("DownloadFiles", controller.HandleDownloadMessage)
    go pubsub.Subscribe("FileUploaded", controller.HandleUploadMessage)

	// cookie 验证
	r.Use(utils.AuthRequired())
	r.POST("/api/auth", controller.Auth)
	r.POST("/api/share", controller.Share)
	r.POST("/api/listFiles", controller.ListFiles)
	r.POST("/api/deleteFile", controller.DeleteFile)
	r.POST("/api/rename", controller.RenameFile)
	r.GET("/api/download", controller.FileDownload)
	r.POST("/api/upload/:downloadToken", controller.Upload)
	r.POST("/api/mkdir", controller.Mkdir)
	r.Run(":5092")
}
