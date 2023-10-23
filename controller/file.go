package controller

import (
	"encoding/json"
    "strconv"
	"strings"
	"time"
    "path"

    "github.com/gin-gonic/gin"

	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/model"
	"git-sec.com/pandownloader/module/utils"
)

// 请求参数
type TokenJson struct {
	PathHash   string   `json:"path_hash"`
	FileId     int      `json:"file_id"`
	PathHashes []string `json:"path_hashes"`
	FileName   string   `json:"file_name"`
}

/**
 * 获取文件列表
 */
func ListFiles(c *gin.Context) {
	var (
		jsonParam TokenJson
	)
	token := c.Request.Header.Get("P-Token")
	if token == "" {
		c.JSON(200, StatusInvalidParameter)
		return
	}

	// 参数校验
	err := c.ShouldBindJSON(&jsonParam)
	if err != nil {
		c.JSON(200, StatusInvalidParameter)
		return
	}
	// 获取分享数据
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, StatusInvalidToken)
		return
	}
	// 根据parentId获取子文件和子文件夹
	parentId := share.ItemSource
	logger.GetLogger().Printf("Get files: %s, Item: %d, PathHash: %s", token, share.ItemSource, jsonParam.PathHash)
	// 获取storage数据
	storage, err := model.GetStorage("home::" + share.Owner)
	if err != nil {
		c.JSON(200, StatusStorageNotFound)
		return
	}
	// 获取子文件夹数据
	if jsonParam.PathHash != "" {
		// 获取该文件数据
		file, err := model.GetFile(storage.StorageId, jsonParam.PathHash)
		if err != nil {
			c.JSON(200, StatusFileNotFound)
			return
		}
		// 判断是否为文件夹
		if file.MimeType != 2 {
			c.JSON(200, StatusFileNotFound)
			return
		}
		parentId = file.FileId
	}
	// 获取文件列表
	filesItem, err := model.GetFileList(storage.StorageId, parentId)
	if err != nil {
		c.JSON(200, StatusNoData)
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"total":   len(filesItem),
		"data":    filesItem,
	})
}

/**
 * 删除文件，权限为31
 */
func DeleteFile(c *gin.Context) {
	var jsonParam TokenJson
	// 参数校验
	token := c.Request.Header.Get("P-Token")
	if token == "" {
		c.JSON(200, StatusInvalidAuthParameter)
		return
	}
	err := c.ShouldBindJSON(&jsonParam)
	if err != nil {
		c.JSON(200, StatusInvalidParameter)
		return
	}
	// 参数校验
	if len(jsonParam.PathHashes) <= 0 {
		c.JSON(200, StatusInvalidParameter)
		return
	}
	// 获取分享数据
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, StatusInvalidToken)
		return
	}
	// 判断权限
	if share.Permissions & model.OC_PERMISSION_DELETE == 0 {
		c.JSON(200, StatusNoPermission)
		return
	}
	// 获取storage数据
	storage, err := model.GetStorage("home::" + share.Owner)
	if err != nil {
		c.JSON(200, StatusStorageNotFound)
		return
	}
	// 多文件删除
	for _, v := range jsonParam.PathHashes {
		// 获取该文件数据
		file, err := model.GetFile(storage.StorageId, utils.Md5V2(v))
		if err != nil {
			c.JSON(200, StatusFileNotFound)
			return
		}
		// 删除并移动文件至回收站
		model.SetTrashBin(file, share.Owner)
		// 添加files_trash记录
		trashModel := model.TrashModel{
			Id:        file.Name,
			User:      share.Owner,
			Timestamp: time.Now().Unix(),
			Location:  share.FileTarget,
		}
		trashId, _ := model.AddTrash(trashModel)
		// 添加动态
		subject := map[string]interface{}{
			strconv.Itoa(file.FileId): file.Path,
		}
		subjectParams, err := json.Marshal(subject)
		activityModel := model.ActivityModel{
			Timestamp:     time.Now().Unix(),
			Priority:      30,
			Type:          "file_deleted",
			User:          "",
			AffectedUser:  share.Owner,
			App:           "files",
			Subject:       "deleted_by",
			SubjectParams: "[" + string(subjectParams) + "]",
			MessageParams: "[]",
			File:          strings.Replace(file.Path, "files", "", 1),
			ObjectType:    "files",
			ObjectId:      file.FileId,
		}
		activityId, _ := model.AddActivity(activityModel)
		logger.GetLogger().Println("trashId", trashId, "activityId", activityId)
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
	})
}

/**
 * 重命名文件，权限为31
 */
func RenameFile(c *gin.Context) {
	var (
		jsonParam TokenJson
	)
	// 参数校验
	token := c.Request.Header.Get("P-Token")
	if token == "" {
		c.JSON(200, StatusInvalidAuthParameter)
		return
	}
	err := c.ShouldBindJSON(&jsonParam)
	if err != nil {
		c.JSON(200, StatusInvalidParameter)
		return
	}
	if jsonParam.PathHash == "" || jsonParam.FileName == "" {
		c.JSON(200, StatusInvalidParameter)
		return
	}

	// 获取分享数据
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, StatusInvalidToken)
		return
	}

	if share.Permissions != 31 {
		c.JSON(200, StatusNoPermission)
		return
	}

	// 获取storage数据
	storage, err := model.GetStorage("home::" + share.Owner)
	if err != nil {
		c.JSON(200, StatusStorageNotFound)
		return
	}

	// 获取该文件数据
	file, err := model.GetFile(storage.StorageId, jsonParam.PathHash)
	if err != nil {
		c.JSON(200, StatusFileNotFound)
		return
	}
	if file.Name == jsonParam.FileName {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "success",
			"data":    file,
		})
		return
	}
    
	// 判断是否为文件夹
    filename := path.Base(jsonParam.FileName)
	model.WalkPath(file.Path, share.Owner, filename, storage.StorageId)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    file,
	})
}
