package model

import (
	"git-sec.com/pandownloader/module/db"
)

/**
 * file_trash model
 */
type TrashModel struct {
	AutoId    int
	Id        string
	User      string
	Timestamp int64
	Location  string
	Type      string
	Mime      string
}

/**
 * 回收站添加
 */
func AddTrash(file TrashModel) (int, error) {
	rs := db.Db.Table("oc_files_trash").Create(&file)
	return file.AutoId, rs.Error
}
