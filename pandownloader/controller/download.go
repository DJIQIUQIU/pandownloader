package controller

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"git-sec.com/pandownloader/elk"
	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/model"
	"git-sec.com/pandownloader/module/db"
	"git-sec.com/pandownloader/module/utils"
)

// 下载选中的文件
type SelectedFiles struct {
	Token      string   `json:"token" binding:"required"`
	PathHashes []string `json:"path_hashes" binding:"required"`
}

// 文件信息
type FileInfo struct {
	Id      int
	Owner   string
	Name    string
	Path    string
	RelPath string
	Size    int
	Empty   bool
}

/* 下载文件 */
func FileDownload(c *gin.Context) {
	token := c.Query("token")
	path := c.QueryArray("path")
	logger.GetLogger().Print("download file:", token, path)
	if token == "" {
		c.JSON(200, gin.H{
			"code": 1000,
			"msg":  "parameter invalid",
		})
		return
	}
	if len(path) < 1 {
		c.JSON(200, gin.H{
			"code": 1001,
			"msg":  "parameter invalid: select nothing",
		})
		return
	}
	// path转换path_hash
	for i, v := range path {
		path[i] = utils.Md5V2(v)
	}
	// 获取分享信息
	share, err := model.GetShare(token)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 2002,
			"msg":  "token invalid",
		})
		return
	}
	storage, err := model.GetStorage("home::" + share.Owner)
	if err != nil {
		logger.GetLogger().Print("[WARN] Cannot find storage: ", share.Owner)
		c.JSON(200, gin.H{
			"code": 2003,
			"msg":  "Cannot find storage",
		})
		return
	}
	panPath, _ := utils.Cfg.GetValue("file", "pan_path")
	if len(path) == 1 {
		file, err := model.GetFile(storage.StorageId, path[0])
		if err != nil {
			logger.GetLogger().Printf("[WARN] Cannot find file %s (%s)", path[0], share.Owner)
			c.JSON(200, gin.H{
				"code": 2004,
				"msg":  "Cannot find file",
			})
			return
		}
		if !file.IsDirectory() {
			singleDownload(c, share, file, panPath)
			return
		}
	}
	archiveDownload(c, share, storage, path, panPath)
}

// 下载单文件
func singleDownload(c *gin.Context, share model.ShareModel, file model.FileModel, panPath string) {
	beforeDownload(file.FileId, share.Owner, c.ClientIP(), file.Path, file.Size)
	filePath := fmt.Sprintf("%s/%s/%s", panPath, share.Owner, file.Path)
	fd, err := os.Open(filePath)
	if err != nil {
		logger.GetLogger().Println("[ERROR] 获取文件失败:", err, filePath)
		c.Redirect(http.StatusFound, "/404")
		return
	}
	defer fd.Close()
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	c.FileAttachment(filePath, file.Name)
	// c.File(filePath)
}

// 下载目录、批量下载, 压缩成 zip
func archiveDownload(c *gin.Context, share model.ShareModel, storage model.StorageModel, pathHashes []string, panPath string) {
	fchan := make(chan FileInfo)
	go func() {
		genFilePath(share, storage, pathHashes, panPath, fchan)
		close(fchan)
	}()
	logger.GetLogger().Printf("Download Shared File: %s (owner: %s, id: %d)", share.FileTarget, share.Owner, share.ItemSource)

	fname := url.PathEscape(strings.Split(strings.TrimPrefix(share.FileTarget, "/"), "/")[0])
	c.Writer.Header().Set("Content-type", "application/zip")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", fname))
	c.Writer.WriteHeader(http.StatusOK)
	err := streamDownload(c, fchan)
	if err != nil {
		conn, _, err := c.Writer.Hijack()
		if err != nil {
			return
		}
		conn.Close()
	}
}

func streamDownload(c *gin.Context, fc chan FileInfo) error {
	w := c.Writer
	zf := zip.NewWriter(w)
	defer zf.Close()
	for fileinfo := range fc {
		fileTmp, err := os.Open(fileinfo.Path)
		if err == nil {
			beforeDownload(fileinfo.Id, fileinfo.Owner, c.ClientIP(), fileinfo.RelPath, fileinfo.Size)
			header := &zip.FileHeader{
				Name:     fileinfo.Name,
				Method:   zip.Store,
				Modified: time.Now(),
			}
			entryWriter, err := zf.CreateHeader(header)
			if err != nil {
				logger.GetLogger().Printf("[Error] Create %s in zipfile: %v", fileinfo.Path, err)
				return err
			}
			_, err = io.Copy(entryWriter, fileTmp)
			if err != nil {
				logger.GetLogger().Printf("[Error] Add %s to zipfile: %v", fileinfo.Path, err)
				return err
			}
			zf.Flush()
			w.Flush()
		} else {
			logger.GetLogger().Print("Open file error", fileinfo.Path)
		}
	}
	return nil
}

// 生成下载的文件路径
func genFilePath(share model.ShareModel, storage model.StorageModel, paths []string, panPath string, fc chan FileInfo) {
	for _, pathHash := range paths {
		logger.GetLogger().Print("Get File:", pathHash)
		file, err := model.GetFile(storage.StorageId, pathHash)
		if err != nil {
			logger.GetLogger().Printf("[WARN] Cannot find file %s (%s)", pathHash, share.Owner)
			continue
		}
		if file.IsDirectory() {
			files, err := model.GetFileList(file.Storage, file.FileId)
			logger.GetLogger().Println("SubPaths:", files)
			if len(files) < 1 && err != nil {
				continue
			}
			subpaths := make([]string, len(files))
			for i, f := range files {
				subpaths[i] = f.PathHash
			}
			genFilePath(share, storage, subpaths, panPath, fc)
			continue
		}
		fpath := panPath + "/" + share.Owner + "/" + file.Path
		fname := strings.TrimPrefix(strings.TrimPrefix(file.Path, "/"), "files")
		logger.GetLogger().Printf("Add File to Zip: %s %s", fname, fpath)
		fc <- FileInfo{
			Id:      file.FileId,
			Owner:   share.Owner,
			Name:    fname,
			Path:    fpath,
			RelPath: file.Path,
			Size:    file.Size,
			Empty:   false,
		}
	}
}

type DownloadFileData struct {
	FileId   int    `json:"fileid"`
	User     string `json:"user"`
	IP       string `json:"ip"`
	FilePath string `json:filepath`
	FileSize int    `json:size`
}

// 下载文件前, 记录下载文件信息到 ELK
func beforeDownload(fileId int, user string, ip string, filePath string, size int) error {
	record := &model.FileDownload{
		FileId:      fileId,
		FilePath:    filePath,
		CreatedAt:   time.Now(),
		CreatedUser: user,
		IP:          ip,
		FileSize:    size,
	}
	record.Save()

	message := DownloadFileData{
		FileId:   fileId,
		User:     user,
		IP:       ip,
		FilePath: filePath,
		FileSize: size,
	}
	data, err := json.Marshal(&message)
	if err != nil {
		logger.GetLogger().Println("Publish `DownloadFiles` Error:", message, err)
		return err
	}

	db.Publish("DownloadFiles", string(data))
	return nil
}

func HandleDownloadMessage(message []byte) error {
	var data DownloadFileData
	if err := json.Unmarshal(message, &data); err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/%s|%s", data.User, data.FilePath, data.User)
	elkRecord := elk.NewDownloadFileRecord(
		data.FileId,
		data.User,
		data.IP,
		path,
		data.FileSize)
	elkRecord.Send()
	return nil
}
