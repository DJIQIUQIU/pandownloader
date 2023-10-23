package model

import (
	"git-sec.com/pandownloader/module/db"
)

var (
    OC_PERMISSION_NONE = 0
    OC_PERMISSION_READ = 1
    OC_PERMISSION_UPDATE = 2
    OC_PERMISSION_CREATE = 4
    OC_PERMISSION_DELETE = 8
    OC_PERMISSION_SHARE = 16
    OC_PERMISSION_ALL = 31
)
/**
 * share Model
 */
type ShareModel struct {
	Id          int
	ItemType    string
	ItemSource  int
	Password    string
	Owner       string `gorm:"column:uid_owner"`
	FileTarget  string
	Permissions int
	Token       string
}

/**
 * 获取分享数据
 */
func GetShare(token string) (ShareModel, error) {
	var (
		share ShareModel
	)
	result := db.Db.Table("oc_share").First(&share, "token = ?", token)
	return share, result.Error
}
