package model

import (
	"fmt"
    "os"
    "path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/module/db"
	"git-sec.com/pandownloader/module/utils"
)

/**
 * fileCache model
 */
type FileModel struct {
	FileId          int `gorm:"column:fileid;primary_key"`
	Storage         int
	Path            string
	PathHash        string
	ParentId        int `gorm:"column:parent"`
	Name            string
	MimeType        int `gorm:"column:mimetype"`
	MimePart        int `gorm:"column:mimepart"`
	Size            int
	Mtime           int64
	StorageMtime    int64
	Encrypted       int
	UnencryptedSize int
	Etag            string
	Permissions     int
	Checksum        string
	MimeTypeStr     string `gorm:"-"`
}

/**
 * 根据父ID获取文件列表
 */
func GetFileList(storageId int, parentId int) ([]FileModel, error) {
	var (
		fileModels []FileModel
	)
	rows := db.Db.Table("oc_filecache").Where("storage =? and parent = ?", storageId, parentId).Find(&fileModels)
	if rows.Error != nil {
		logger.GetLogger().Println("GetMimeType err:", rows.Error)
	} else {
		SetMimeTypes()
		for k, file := range fileModels {
			fileModels[k].MimeTypeStr = db.HGet("mimeType", strconv.Itoa(file.MimeType)).Val()
		}
	}
	return fileModels, rows.Error
}

/**
 * 获取单文件数据
 */
func GetFileById(fileId int) (FileModel, error) {
	var (
		file FileModel
	)
	row := db.Db.Table("oc_filecache").First(&file, "fileid = ?", fileId)
	if row.Error != nil {
		logger.GetLogger().Println("GetFile err:", row.Error)
	}
	return file, row.Error
}

/**
 * 获取单文件数据
 */
func GetFile(StorageId int, pathHash string) (FileModel, error) {
	var (
		file FileModel
	)
	row := db.Db.Table("oc_filecache").First(&file, "storage = ? and path_hash = ?", StorageId, pathHash)
	if row.Error != nil {
		logger.GetLogger().Println("GetFile err:", row.Error)
	}
	return file, row.Error
}

func (f FileModel) IsFile() bool {
	return f.MimeType != 1 && f.MimeType != 2 && f.MimeType != 29
}

func (f FileModel) IsDirectory() bool {
	return f.MimeType == 2
}

func AddFile(file FileModel) (int, error) {
	rs := db.Db.Table("oc_filecache").Create(&file)
	return file.FileId, rs.Error
}
func UpdateFile(file FileModel) {
	db.Db.Table("oc_filecache").Save(&file)
}

func SetTrashBin(file FileModel, owner string) {
	filesTrashBin := "files_trashbin"
	panPath, _ := utils.Cfg.GetValue("file", "pan_path")
	panPath = panPath + "/" + owner
	pathList := strings.Split(filesTrashBin+"/"+file.Path, "/")
	pathLength := len(pathList)
	var lastValue []string
	for k, v := range pathList {
		fmt.Println("SetTrashBin:k", k)
		if (k + 1) == pathLength {
			dName := ".d" + strconv.Itoa(int(time.Now().Unix()))
            trashValue := filepath.Join(filesTrashBin, file.Path) + dName
            src := filepath.Join(panPath, file.Path)
            dst := filepath.Join(panPath, trashValue)
            logger.GetLogger().Printf("Trash: %p, FilePath: %p\n", filepath.Join(panPath, trashValue), filepath.Join(panPath, file.Path))
			if utils.Move(src, dst) {
				lastFile, _ := GetFile(file.Storage, utils.Md5V2(strings.Join(lastValue, "/")))
				fmt.Println("lastFile", lastFile)
				file.ParentId = lastFile.FileId
				file.Path = trashValue
				file.PathHash = utils.Md5V2(trashValue)
				file.Name = file.Name + dName
				UpdateFile(file)
			}
		} else {
			last := lastValue
			lastValue = append(lastValue, v)
			fmt.Println("mkdir:", strings.Join(lastValue, "/"))
			if utils.Mkdir(panPath + "/" + strings.Join(lastValue, "/")) {
				lastFile, _ := GetFile(file.Storage, utils.Md5V2(strings.Join(last, "/")))
				fileModel := FileModel{
					Storage:         file.Storage,
					Path:            strings.Join(lastValue, "/"),
					PathHash:        utils.Md5V2(strings.Join(lastValue, "/")),
					ParentId:        lastFile.FileId,
					Name:            v,
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
				AddFile(fileModel)
			}
		}
	}
}

/**
 * 目录遍历
 */
func WalkPath(filePath string, owner string, newName string, storageId int) {
	panPath, _ := utils.Cfg.GetValue("file", "pan_path")
    sourceFile := filepath.Join(panPath, owner, filePath)
    if !utils.Exists(sourceFile) {
		return
	}
    newName = path.Base(newName)
	oldPath := strings.Split(filePath, "/")
	oldPath = oldPath[:len(oldPath)-1]
    newPath := filepath.Join(panPath, owner, strings.Join(append(oldPath, newName), "/"))

	index := -1
	//获取当前目录下的所有文件或目录信息
	filepath.Walk(sourceFile, func(path string, info os.FileInfo, err error) error {
		currentPath := strings.Replace(path, panPath+"/"+owner+"/", "", 1)
		currentSplit := strings.Split(currentPath, "/")
		if currentPath == filePath {
			index = len(currentSplit) - 1
		}
		if index == -1 {
			return nil
		}
		currentSplit[index] = newName
		newPathString := strings.Join(currentSplit, "/")

		pathHash := utils.Md5V2(currentPath)
		file, _ := GetFile(storageId, pathHash)
		file.Path = newPathString
		file.PathHash = utils.Md5V2(newPathString)
		// 重命名文件
		if currentPath == filePath {
			file.Name = newName
		}
		fmt.Println("file:", file, " newPathString:", newPathString, " pathHash:", pathHash)
		UpdateFile(file)
		// fmt.Println(strings.Replace(path, panPath+"/"+owner+"/", "", 1)) //打印path信息
		// fmt.Println(info.Name())                                         //打印文件或目录名
		// fmt.Println(newPath)                                             //打印文件或目录名
		// fmt.Println("error", err)
        logger.GetLogger().Printf("%v\n", err)
		return nil
	})
	utils.Move(sourceFile, newPath)
}
