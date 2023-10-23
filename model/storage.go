package model

import (
	"git-sec.com/pandownloader/logger"
	"git-sec.com/pandownloader/module/db"
)

/**
 * storage model
 */
type StorageModel struct {
	StorageId int    `gorm:"column:numeric_id"`
	Username  string `gorm:"column:id"`
	Available int
}

/**
 * 获取storage数据
 */
func GetStorage(id string) (StorageModel, error) {
	var (
		storage StorageModel
	)
	row := db.Db.Table("oc_storages").First(&storage, "available = 1 and id = ?", id)
	if row.Error != nil {
		logger.GetLogger().Println("GetStorage err:", row.Error)
	}
	return storage, row.Error
}
