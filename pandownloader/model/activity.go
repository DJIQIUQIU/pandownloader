package model

import (
	"git-sec.com/pandownloader/module/db"
)

/**
 * activity model
 */
type ActivityModel struct {
	ActivityId    int
	Timestamp     int64
	Priority      int
	Type          string
	User          string
	AffectedUser  string `gorm:"column:affecteduser"`
	App           string
	Subject       string
	SubjectParams string `gorm:"column:subjectparams"`
	MessageParams string `gorm:"column:messageparams"`
	File          string
	Link          string
	ObjectType    string
	ObjectId      int
}

/**
 * 动态添加
 */
func AddActivity(activity ActivityModel) (int, error) {
	rs := db.Db.Table("oc_activity").Create(&activity)
	return activity.ActivityId, rs.Error
}
