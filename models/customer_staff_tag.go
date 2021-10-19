package models

import (
	"gorm.io/gorm/clause"
	"msg/constants"
)

type CustomerStaffTag struct {
	ExtCorpModel
	CustomerStaffID string `json:"customer_staff_id" gorm:"type:bigint;index;"`
	// TagID 标签id
	ExtTagID string `json:"ext_tag_id"`
	// GroupName 该成员添加此外部联系人所打标签的分组名称（标签功能需要企业微信升级到2.7.5及以上版本）
	GroupName string `json:"group_name"`
	// TagName 该成员添加此外部联系人所打标签名称
	TagName string `json:"tag_name"`
	// Type 该成员添加此外部联系人所打标签类型, 1-企业设置, 2-用户自定义
	Type constants.FollowUserTagType `gorm:"type:tinyint" json:"type"`
	Timestamp
}

// CreateInBatches 批量创建

func (c CustomerStaffTag) CreateInBatches(customerStaffTags []CustomerStaffTag) error {
	return DB.CreateInBatches(customerStaffTags, len(customerStaffTags)).Error
}

func (c CustomerStaffTag) Delete(customerStaffId string, extTagsIDs []string) error {
	return DB.Where("customer_staff_id = ?", customerStaffId).
		Where("tag_id in (?)", extTagsIDs).
		Delete(&CustomerStaffTag{}).Error
}

func (c CustomerStaffTag) Upsert(tag []CustomerStaffTag) error {
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "customer_staff_id"}},
		DoUpdates: clause.AssignmentColumns(
			[]string{"group_name", "ext_tag_id", "type", "tag_name"},
		),
	}).CreateInBatches(&tag, len(tag)).Error
}
