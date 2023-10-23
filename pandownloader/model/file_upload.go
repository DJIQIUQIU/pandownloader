package model

import (
	"git-sec.com/pandownloader/module/db"
)

/**
 * fileupload model
 */
type FileUploaded struct {
    Id int `gorm:"column:id;primary_key"`
    FileId int `gorm:"column:fileid"`
    User string
    Created int
    IP string
    Deleted int
}


func (record *FileUploaded) AddExtendInfo() (int, error) {
    rs := db.Db.Table("oc_filecache_uploaded").Create(record)
    return record.Id, rs.Error
}


