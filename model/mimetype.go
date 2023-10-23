package model

import (
	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/module/db"
	"strconv"
)

/**
 * mimetype model
 */
type MimeTypeModel struct {
	Id       int
	MimeType string `gorm:"column:mimetype"`
}

/**
 * 获取mimeType数据
 */
func SetMimeTypes() map[string]string {
	var (
		mimeTypes []MimeTypeModel
	)
	mimeTypesResult := db.HGetAll("mimeType").Val()
    if len(mimeTypesResult) < 1 || mimeTypesResult == nil {
		rows := db.Db.Table("oc_mimetypes").Select("id", "mimetype").Find(&mimeTypes)
		if rows.Error != nil {
			logger.GetLogger().Println("err", rows.Error)
		}
		mimeTypesResult = make(map[string]string)
		for _, mimeType := range mimeTypes {
			db.HSet("mimeType", mimeType.Id, mimeType.MimeType)
			mimeTypesResult[strconv.Itoa(mimeType.Id)] = mimeType.MimeType
		}
	}
	return mimeTypesResult
}

func GetMimeTypeId(mimeType string) int {
	mimeTypes := SetMimeTypes()
	for k, v := range mimeTypes {
		if mimeType == v {
			k, _ := strconv.Atoi(k)
			return k
		}
	}
	return 21
}

func GetMimeTypeStr(id int) string {
	SetMimeTypes()
	return db.HGet("mimeType", strconv.Itoa(id)).Val()
}
