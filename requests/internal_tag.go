package requests

import "msg/common/app"

type CreateInternalTagReq struct {
	Name string `json:"name" validate:"required"`
}

type DeleteInternalTagReq struct {
	IDs []string `json:"ids" validate:"required,gt=0"`
}

type QueryInternalTagReq struct {
	ExtStaffID string `json:"ext_staff_id" form:"ext_staff_id" validate:"required"`
	app.Sorter
	app.Pager
}
