package controller

import (
	"encoding/json"
	// "fmt"
	"path"
    "path/filepath"
	"strconv"
	"strings"
	"time"
    // "os"

	"github.com/gin-gonic/gin"

	"git-sec.com/pandownloader/elk"
	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/model"
	"git-sec.com/pandownloader/module/db"
	"git-sec.com/pandownloader/module/utils"
)

type mkdirJson struct {
	Name string `json:"name" binding:"required"`
}

/**
 * 文件上传, 权限不能为4
 */
func Upload(c *gin.Context) {
	// 上传参数
	// 参数校验
	token := c.Request.Header.Get("P-Token")
    form, _ := c.MultipartForm()
	files := form.File["file"]
    if token == "" || len(files) <= 0 {
		c.JSON(200, StatusInvalidParameter)
		return
	}
	// 查询share
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, StatusTokenNotFound)
		return
	}
    // 权限判断
	if share.Permissions & model.OC_PERMISSION_CREATE == 0 {
		c.JSON(200, StatusNoPermission)
		return
	}

    // 获取分享根目录数据
	file, err := model.GetFileById(share.ItemSource)
	if err != nil {
		c.JSON(200, StatusFileNotFound)
		return
	}
    // 是否子目录上传
	urlPath := c.Query("path")
    logger.GetLogger().Printf("Path: %p\n", urlPath)
	newPath := file.Path
	if urlPath != "" {
        urlPath, _ = filepath.Abs(urlPath)
		newPath = filepath.Join(newPath, urlPath)
		// 重新获取父目录文件信息
		file, err = model.GetFile(file.Storage, utils.Md5V2(newPath))
        if err != nil {
            c.JSON(200, StatusInvalidParameter)
            return 
        }
	}

	// nextcloud文件存储路径
    panPath, _ := utils.Cfg.GetValue("file", "pan_path")
    dst := filepath.Join(panPath, share.Owner, file.Path)
	var fileModels []model.FileModel
	for _, f := range files {
        logger.GetLogger().Printf("Upload: %p, Type: %s", f.Filename, f.Header["Content-Type"])
        safeFilename := path.Base(f.Filename)
        target := filepath.Join(dst, safeFilename)
		// 是否需要重命名
		target = utils.ReName(target)
        safeFilename = path.Base(target)
		// 保存文件
		err := c.SaveUploadedFile(f, target)
		if err != nil {
			logger.GetLogger().Printf("Upload Error: %v\n", err)
            continue
		}
		// 写入DB
        targetPath := filepath.Join(newPath, safeFilename)
		pathHash := utils.Md5V2(targetPath)
		fileModel := model.FileModel{
			Storage:         file.Storage,
			Path:            targetPath,
			PathHash:        pathHash,
			ParentId:        file.FileId,
			Name:            safeFilename,
			MimeType:        model.GetMimeTypeId(f.Header["Content-Type"][0]),
			MimePart:        1,
			Size:            int(f.Size),
			Mtime:           time.Now().Unix(),
			StorageMtime:    time.Now().Unix(),
			Encrypted:       0,
			UnencryptedSize: 0,
			Etag:            "0",
			Permissions:     31,
			Checksum:        "",
		}
		fileId, err := model.AddFile(fileModel)
		fileModel.MimeTypeStr = model.GetMimeTypeStr(fileModel.MimeType)
		fileModel.FileId = fileId
		fileModels = append(fileModels, fileModel)
		subject := map[string]interface{}{
			strconv.Itoa(file.FileId): file.Path,
		}
		subjectParams, err := json.Marshal(subject)
		// 写入动态
		activityModel := model.ActivityModel{
			Timestamp:     time.Now().Unix(),
			Priority:      30,
			Type:          "file_created",
			User:          "",
			AffectedUser:  share.Owner,
			App:           "files",
			Subject:       "created_public",
			SubjectParams: "[" + string(subjectParams) + "]",
			MessageParams: "[]",
			File:          strings.Replace(file.Path, "files", "", 1),
			ObjectType:    "files",
			ObjectId:      file.FileId,
		}
		model.AddActivity(activityModel)
		// 写入扩展表
		AfterUpload(fileId, "", c.ClientIP(), fileModel.Path)
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    fileModels,
	})
	return
}

/**
 * 文件夹创建
 */
func Mkdir(c *gin.Context) {
	// 参数
	token := c.Request.Header.Get("P-Token")
	if token == "" {
        c.JSON(200, StatusInvalidParameter)
		return
	}
    // 查询share
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, StatusTokenNotFound)
		return
	}
    // 权限判断
	if share.Permissions & model.OC_PERMISSION_CREATE == 0 {
		c.JSON(200, StatusNoPermission)
		return
	}

    var mkdirJson mkdirJson
	err = c.ShouldBindJSON(&mkdirJson)
	if err != nil {
		c.JSON(200, StatusInvalidParameter)
		return
	}

	// 获取分享根目录数据
	file, err := model.GetFileById(share.ItemSource)
	if err != nil {
		c.JSON(200, StatusFileNotFound)
		return
	}
	// 是否子目录上传
    urlPath := c.Query("path")
    basePath := file.Path
	if urlPath != "" {
        urlPath, _ = filepath.Abs(urlPath)
        basePath = filepath.Join(basePath, urlPath)
		// 重新获取父目录文件信息
		file, err = model.GetFile(file.Storage, utils.Md5V2(basePath))
        if err != nil {
            c.JSON(200, StatusInvalidParameter)
            return
        }
	}
	// nextcloud文件存储路径
	panPath, _ := utils.Cfg.GetValue("file", "pan_path")
    folderName, err := filepath.Abs(mkdirJson.Name)
    folderName = path.Base(folderName)
    dst := filepath.Join(panPath, share.Owner, basePath, folderName)
    logger.GetLogger().Printf("New Folder: %p -> %p\n", mkdirJson.Name, folderName)
	// 是否需要重命名
	dst = utils.ReName(dst)
	fileName := path.Base(dst)
	// 创建文件夹
	if !utils.Mkdir(dst) {
        c.JSON(200, StatusInvalidParameter)
        return
    }
	// 写入DB
    filePath := filepath.Join(basePath, fileName)
	pathHash := utils.Md5V2(filePath)
	fileModel := model.FileModel{
		Storage:         file.Storage,
		Path:            filePath,
		PathHash:        pathHash,
		ParentId:        file.FileId,
		Name:            fileName,
		MimeType:        2,
		MimePart:        1,
		Size:            0,
		Mtime:           time.Now().Unix(),
		StorageMtime:    time.Now().Unix(),
		Encrypted:       0,
		UnencryptedSize: 0,
		Etag:            "0",
		Permissions:     31,
		Checksum:        "",
	}
	fileId, err := model.AddFile(fileModel)
	fileModel.MimeTypeStr = model.GetMimeTypeStr(fileModel.MimeType)
	fileModel.FileId = fileId
	// 写入动态
	subject := map[string]interface{}{
		strconv.Itoa(file.FileId): file.Path,
	}
	subjectParams, err := json.Marshal(subject)
	activityModel := model.ActivityModel{
		Timestamp:     time.Now().Unix(),
		Priority:      30,
		Type:          "file_created",
		User:          "",
		AffectedUser:  share.Owner,
		App:           "files",
		Subject:       "created_public",
		SubjectParams: "[" + string(subjectParams) + "]",
		MessageParams: "[]",
		File:          strings.Replace(file.Path, "files", "", 1),
		ObjectType:    "files",
		ObjectId:      file.FileId,
	}
	model.AddActivity(activityModel)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    fileModel,
	})
	return
}

type UploadFileData struct {
	FileId   int    `json:"fileid"`
	User     string `json:"user"`
	IP       string `json:"ip"`
	FilePath string `json:filepath`
	FileSize int    `json:size`
}

// 上传成功后, 写入文件的扩展信息至 DB, 发布事件
// 扩展信息包括 用户、ip
func AfterUpload(fileId int, user string, ip string, filePath string) error {
	record := &model.FileUploaded{
		FileId:  fileId,
		User:    user,
		IP:      ip,
		Created: int(time.Now().Unix()),
		Deleted: 0,
	}
	record.AddExtendInfo()

	message := UploadFileData{
		FileId:   fileId,
		User:     user,
		IP:       ip,
		FilePath: filePath,
	}
	data, err := json.Marshal(&message)
	if err != nil {
		logger.GetLogger().Println("Publish `FileUploaded` Error:", message, err)
		return err
	}
	db.Publish("FileUploaded", string(data))
	return nil
}

func HandleUploadMessage(message []byte) error {
	var data UploadFileData
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}
	elkRecord := elk.NewFileUploadedRecord(
		data.FileId,
		data.User,
		data.IP,
		data.FilePath,
	)
	elkRecord.Send()
	return nil
}
