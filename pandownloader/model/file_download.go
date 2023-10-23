package model

import (
    "time"
	"git-sec.com/pandownloader/module/db"
)

/**
 * fileupload model
 */
type FileDownload struct {
    Id int `gorm:"column:id;primary_key"`
    FileId int `gorm:"column:fileid"`
    FilePath string `gorm:"column:filepath"`
    CreatedAt time.Time `gorm:"column:created_at"`
    CreatedUser string `gorm:"column:created_user"`
    IP string `gorm:"column:ip"`
    FileSize int `gorm:"column:filesize"`
}


func (record *FileDownload) Save() (int, error) {
    rs := db.Db.Table("oc_download_tracker").Create(record)
    return record.Id, rs.Error
}
