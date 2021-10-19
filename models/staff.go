package models

import (
	"fmt"
	redis2 "github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"msg/common/app"
	"msg/common/log"
	"msg/common/redis"
	"msg/common/util"
	"msg/conf"
	"msg/constants"
	"msg/requests"
	"os"
	"time"
)

type Staff struct {
	Model
	// ExtCorpID 外部企业ID
	ExtCorpID string `json:"ext_corp_id" gorm:"index;uniqueIndex:idx_ext_corp_id_ext_staff_id;type:char(18);comment:外部企业ID"`
	//企业内必须唯一。不区分大小写，长度为1~64个字节
	ExtID string `gorm:"type:char(32);uniqueIndex:idx_ext_corp_id_ext_staff_id;comment:外部员工ID" json:"ext_staff_id"`
	// RoleID 角色ID
	RoleID string `json:"role_id" gorm:"type:bigint;comment:角色ID"`
	// RoleType 角色类型
	RoleType string `json:"role_type" gorm:"index;default:staff;comment:'角色类型'" validate:"oneof=superAdmin admin departmentAdmin staff"`
	// 成员名称
	Name string `gorm:"type:varchar(255);comment:员工名" json:"name"`
	// 地址
	Address string `gorm:"type:varchar(255);comment:地址" json:"address"`
	// 别名
	Alias string `gorm:"type:varchar(255);comment:别名" json:"alias"`
	// 头像url
	AvatarURL string `gorm:"type:varchar(128);comment:头像地址" json:"avatar_url"`
	// 邮箱，第三方仅通讯录应用可获取
	Email string `gorm:"type:varchar(128)" json:"email"`
	// 性别
	Gender UserGender `gorm:"type:tinyint;comment:0表示未定义，1表示男性，2表示女性" json:"gender"`
	// 激活状态
	Status constants.UserStatus `gorm:"type:tinyint;comment:激活状态: 1=已激活，2=已禁用，4=未激活，5=退出企业。已激活代表已激活企业微信或已关注微工作台（原企业号）。未激活代表既未激活企业微信又未关注微工作台（原企业号）。" json:"status"`
	// 手机号码
	Mobile string `gorm:"index;type:varchar(11);comment:手机号;" json:"mobile"`
	// 员工个人二维码；第三方仅通讯录应用可获取
	QRCodeURL string `gorm:"type:varchar(255);comment:二维码" json:"qr_code_url"`
	// Telephone 座机；第三方仅通讯录应用可获取
	Telephone string `gorm:"type:char(11);comment:电话" json:"telephone"`
	// IsEnabled 成员的启用状态 0-禁用 1-启用
	Enable int `gorm:"type:tinyint unsigned" json:"enable"`
	// sha1 hash
	Signature string `gorm:"type:char(128);comment:微信返回的内容签名" json:"signature"`
	// 职务信息
	ExternalPosition string `json:"external_position"`
	// 成员对外属性
	ExternalProfile string `json:"external_profile"`
	// 扩展属性
	Extattr string `json:"extattr"`
	// 客户数量
	CustomerCount int `json:"external_user_count"`
	//所属部门ids
	DeptIds     constants.Int64ArrayField `gorm:"type:json" json:"dept_ids"`
	Departments []Department              `gorm:"many2many:StaffDepartment;" json:"departments"`
	// 欢迎语id
	WelcomeMsgID *string `gorm:"type:bigint;index" json:"welcome_msg_id"`
	// 是否授权 1-是 2-否
	IsAuthorized constants.Boolean `gorm:"type:tinyint unsigned" json:"is_authorized"`
	// 开启绘画存档 1-是 2-否
	EnableMsgArch constants.Boolean `gorm:"type:tinyint unsigned;default:2" json:"enable_msg_arch"`
	Timestamp
}

func (s Staff) TableName() string {
	return "staff"
}

// StaffMainInfo 员工的主要信息
type StaffMainInfo struct {
	ID          string           `json:"id"`
	ExtID       string           `json:"ext_id"`
	AvatarURL   string           `json:"avatar_url"`
	RoleType    string           `json:"role_type"`
	RoleID      string           `json:"role_id"`
	Name        string           `json:"name"`
	Departments []MainDepartment `gorm:"many2many:StaffDepartment;" json:"departments"`
}

// StaffsMainInfoCache 员工的主要信息的缓存
type StaffsMainInfoCache struct {
	Staffs []StaffMainInfo `json:"staffs"`
	Total  int64           `json:"total"`
}

// MainDepartment 员工的主要信息中的部门信息
type MainDepartment struct {
	// 企业微信部门id
	ExtID int64 `gorm:"type:int;uniqueIndex:idx_ext_corp_id_ext_dept_id;comment:企微定义的部门ID" json:"ext_id"`
	// 部门名称
	Name string `gorm:"type:varchar(255);comment:部门名称" json:"name"`
	// 上级部门id
	ExtParentID int64 `gorm:"type:int unsigned;comment:上级部门ID,根部门为1" json:"ext_parent_id"`
}

func (s Staff) Get(extStaffID string, extCorpID string, withDepartments bool) (*Staff, error) {
	staff := &Staff{}
	db := DB.Model(&Staff{}).Where("ext_id = ? ", extStaffID)

	if extCorpID != "" {
		db = db.Where("ext_corp_id = ?", extCorpID)
	}
	if withDepartments {
		db = db.Preload("Departments")
	}
	err := db.First(staff).Error

	if err != nil {
		err = errors.WithStack(err)
		return nil, err
	}

	return staff, nil
}

func (s Staff) Query(staff Staff, extCorpID string, sorter *app.Sorter, pager *app.Pager) ([]*Staff, int64, error) {
	staffs := make([]*Staff, 0)
	var total int64
	db := DB.Model(&Staff{}).Where("ext_corp_id = ?", extCorpID)
	if staff.Name != "" {
		db = db.Where("name like ?", staff.Name+"%")
	}

	if len(staff.DeptIds) > 0 {
		db = db.Where("json_contains(dept_ids, (?) )", staff.DeptIds)
	}

	if staff.RoleID != "" {
		db = db.Where("role_id =?", staff.RoleID)
	}

	if staff.RoleType != "" {
		db = db.Where("role_type =?", staff.RoleType)
	}

	if staff.EnableMsgArch == 1 || staff.EnableMsgArch == 2 {
		db = db.Where("enable_msg_arch = ?", staff.EnableMsgArch)
	}

	err := db.Count(&total).Error
	if err != nil || total == 0 {
		err = errors.Wrap(err, "Count staff_event failed")
		return nil, 0, err
	}
	sorter.SetDefault()
	db = db.Order(
		clause.OrderByColumn{
			Column: clause.Column{Name: string(sorter.SortField)},
			Desc:   sorter.SortType == constants.SortTypeDesc,
		})

	pager.SetDefault()
	db = db.Offset(pager.GetOffset()).Limit(pager.GetLimit())

	err = db.Preload("Departments").Find(&staffs).Error
	if err != nil {
		err = errors.Wrap(err, "Find staffs failed")
		return nil, 0, err
	}
	return staffs, total, nil
}

func (s Staff) Upsert(staff Staff) error {
	tx := DB.Begin()
	defer tx.Rollback()
	err := DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "ext_corp_id"}, {Name: "ext_id"}},
		DoUpdates: clause.AssignmentColumns(
			[]string{`extattr`, `external_profile`, `external_position`, `telephone`, `qr_code_url`,
				`mobile`, `status`, `gender`, `email`, `avatar_url`, `alias`, `address`, `name`}),
	}).Create(&staff).Error
	if err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s Staff) EnableInBatches(enableIDs []string, disableIDs []string, extCorpID string) error {
	if len(enableIDs) > 0 {
		return DB.Model(&Staff{}).
			Where("ext_corp_id = ?", extCorpID).
			Where("ext_id  in (?)", enableIDs).
			Update("enable", 1).Error
	}
	if len(disableIDs) > 0 {
		return DB.Model(&Staff{}).
			Where("ext_corp_id = ?", extCorpID).
			Where("ext_id in (?)", enableIDs).
			Update("enable", 0).Error
	}
	return nil
}

func (s Staff) CleanCache(extCorpID string) (err error) {
	keys := fmt.Sprintf(constants.CacheMainStaffInfoKeyPrefix, extCorpID)
	log.Sugar.Debugw("args", "prefix", keys)
	err = redis.RedisClient.Eval(context.TODO(), constants.DelCacheMainStaffInfoKeyScripts, []string{"KEYS"}, keys).Err()
	if errors.Is(err, redis2.Nil) {
		return nil
	}
	return
}

func (s Staff) CachedQueryMainInfo(req requests.QueryMainStaffInfoReq, extCorpID string, pager *app.Pager) (StaffsMainInfoCache, error) {
	var staffsCached StaffsMainInfoCache
	err := redis.GetOrSetFunc(
		fmt.Sprintf(constants.CacheMainStaffInfoKey, extCorpID, req.ExtDepartmentID, pager.GetOffset(), pager.GetLimit()),
		func() (interface{}, error) {
			return s.QueryMainInfo(req, extCorpID, pager)
		},
		time.Hour*24,
		&staffsCached,
	)
	return staffsCached, err
}

func (s Staff) QueryMainInfo(req requests.QueryMainStaffInfoReq, extCorpID string, pager *app.Pager) (res StaffsMainInfoCache, err error) {

	db := DB.Table("staff").
		Joins("left join staff_department sd on sd.staff_id = staff.id").
		Joins("left join department on sd.department_id = department.id").
		Where("staff.ext_corp_id = ? ", extCorpID)

	if req.ExtStaffID != "" {
		db = db.Where("staff.ext_id = ?", req.ExtStaffID)
	}

	if req.ExtDepartmentID != "" {
		db = db.Where("department.ext_department_id = (?)", req.ExtDepartmentID)
	}

	err = db.Distinct("staff.id").Count(&res.Total).Error
	if err != nil || res.Total == 0 {
		err = errors.Wrap(err, "Count StaffMainInfo failed")
		return res, err
	}
	staffs := make([]Staff, 0)
	IDs := make([]string, 0)

	pager.SetDefault()
	db = db.Offset(pager.GetOffset()).Limit(pager.GetLimit())
	err = db.Pluck("staff.id", &IDs).Error
	if err != nil {
		return res, err
	}

	err = DB.Model(&Staff{}).
		Select("id,ext_id,avatar_url,role_id,role_type,name").
		Where("id in ?", IDs).Preload("Departments").Find(&staffs).Error
	if err != nil {
		err = errors.Wrap(err, "Find StaffMainInfo failed")
		return res, err
	}
	log.Sugar.Debugw("staff main info", "depts", util.JsonEncode(staffs[0].Departments))
	err = copier.CopyWithOption(&res.Staffs, staffs, copier.Option{DeepCopy: true})
	if err != nil {
		return res, err
	}

	return res, err
}

func (s Staff) GetMainInfo(extStaffID string, extCorpID string) (res StaffMainInfo, err error) {

	var staff Staff
	err = DB.Model(&Staff{}).
		Select("id,ext_id,avatar_url,role_id,role_type,name").
		Where("ext_corp_id = ?", extCorpID).
		Where(" ext_id = ?", extStaffID).Preload("Departments").First(&staff).Error
	if err != nil {
		err = errors.Wrap(err, "Find StaffMainInfo failed")
		return res, err
	}
	log.Sugar.Debugw("staff main info", "depts", util.JsonEncode(staff))
	err = copier.CopyWithOption(&res, staff, copier.Option{DeepCopy: true})
	if err != nil {
		return res, err
	}

	log.Sugar.Debugw("staff main info", "depts", util.JsonEncode(res))
	return res, err
}

// UpdateStaffMsgArchStatus 更新员工绘画存档开关
func (s Staff) UpdateStaffMsgArchStatus(extCorpID string, extStaffIDs []string, status constants.Boolean) (err error) {
	return DB.Model(&Staff{}).
		Where("ext_corp_id = ? and  ext_staff_id in (?)", extCorpID, extStaffIDs).
		Update("enable_msg_arch = ?", status).Error
}

// --------- welcome msg  -----------------

func (s Staff) UpdateWelcomeMsg(tx *gorm.DB, extCorpID string, staffID []string, msgID string) error {
	return tx.Model(&Staff{}).
		Where("ext_corp_id = ?", extCorpID).
		Where("ext_id in (?)", staffID).
		Update("welcome_msg_id", msgID).Error
}

func (s Staff) CreateStaffInBatches(newStaffs []Staff) error {
	return DB.Model(&Staff{}).CreateInBatches(newStaffs, len(newStaffs)).Error
}

func (s Staff) GetStaffByIDSAndSignatures(ids, signatures []string) (updatedIDs []string, err error) {
	if err = DB.Model(&Staff{}).
		Where("ext_id in ? and signature not in ?", ids, signatures).
		Pluck("ext_id", &updatedIDs).Error; err != nil {
		err = errors.Wrap(err, "GetStaffByIDSAndSignatures failed")
		return
	}
	return
}

func (s Staff) GetAllStaffIDs() (allUserIds []string, err error) {
	err = DB.Model(&Staff{}).Pluck("ext_id", &allUserIds).Error
	if err != nil {
		err = errors.Wrap(err, "GetAllStaffIDs failed")
		return
	}
	return
}

func (s Staff) UpdateAuthorizedStatus(staffIDs []string) error {
	return DB.Model(&Staff{}).Where("ext_id in (?)", staffIDs).Update("is_authorized", constants.True).Error
}

func (s Staff) RemoveOriginalWelcomeMsg(tx *gorm.DB, welcomeMsgId string) error {
	return tx.Model(&Staff{}).Where("welcome_msg_id = ?", welcomeMsgId).Update("welcome_msg_id", nil).Error
}

func SetupStaffRole() {
	tx := DB.Begin()
	defer tx.Rollback()

	// 清空超级管理员权限
	err := tx.Model(&Staff{}).Where("role_type = ?", conf.Settings.App.SuperAdminPhone).Updates(&Staff{
		RoleType: string(constants.RoleTypeStaff),
		RoleID:   string(constants.DefaultCorpStaffRoleID),
	}).Error
	if err != nil {
		log.TracedError("clean SuperAdmin role failed", errors.WithStack(err))
		os.Exit(1)
	}

	// 根据conf里的SuperAdminPhone配置设置超级管理员员工
	err = tx.Model(&Staff{}).Where("mobile in (?)", conf.Settings.App.SuperAdminPhone).Updates(&Staff{
		RoleType: string(constants.RoleTypeSuperAdmin),
		RoleID:   string(constants.DefaultCorpSuperAdminRoleID),
	}).Error
	if err != nil {
		log.TracedError("set SuperAdmin role failed", errors.WithStack(err))
		os.Exit(1)
	}

	err = tx.Commit().Error
	if err != nil {
		log.TracedError("Commit failed", errors.WithStack(err))
		os.Exit(1)
	}

	err = Staff{}.CleanCache(conf.Settings.WeWork.ExtCorpID)
	if err != nil {
		log.TracedError("CleanCache failed", errors.WithStack(err))
		os.Exit(1)
	}

}
