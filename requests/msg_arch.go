package requests

import (
	"msg/common/app"
	"msg/constants"
)

// QueryChatMsgReq  会话中的消息列表
type QueryChatMsgReq struct {
	// 员工外部ID，from
	ExtStaffID string `json:"ext_staff_id" form:"ext_staff_id" validate:"required"`
	// 接收者id，员工或客户外部ID
	ReceiverID string `json:"receiver_id" form:"receiver_id" validate:"required"`
	// 消息类型
	MsgType string `json:"msg_type" form:"msg_type" validate:"omitempty"`
	// 起止时间
	SendAtStart constants.DateField `json:"send_at_start" form:"send_at_start" validate:"omitempty"`
	SendAtEnd   constants.DateField `json:"send_at_end" form:"send_at_end" validate:"omitempty"`
	// 上下文上限的消息ID
	MaxID string `json:"max_id" form:"max_id" validate:"omitempty"`
	// 上下文下限的消息ID
	MinID string `json:"min_id" form:"min_id" validate:"omitempty"`
	// 上下文条数限制
	Limit int `json:"limit" form:"limit" validate:"omitempty"`
	app.Sorter
	app.Pager
}

// QuerySessionReq 查询会话列表
type QuerySessionReq struct {
	// 员工外部ID
	ExtStaffID string `json:"ext_staff_id" form:"ext_staff_id" validate:"required"`
	// 类型 room-群聊 external-外部 internal-内部
	SessionType string `json:"session_type" form:"session_type" validate:"oneof=room external internal"`
	// 客户名
	Name string `json:"name" form:"name" validate:"omitempty"`
	app.Pager
	app.Sorter
}

type SearchMsgReq struct {
	Keyword    string `json:"keyword" form:"keyword" validate:"required"`
	ExtStaffID string `json:"ext_staff_id" form:"ext_staff_id" validate:"required"`
	ExtPeerID  string `json:"ext_peer_id" form:"ext_peer_id" validate:"required"`
	app.Pager
}

// -------------  req for inner srv --------------

type SyncReq struct {
	ExtCorpID string `json:"ext_corp_id" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}

type InnerQuerySessionsReq struct {
	QuerySessionReq
	ExtCorpID string `json:"ext_corp_id" form:"ext_corp_id" validate:"required"`
	Signature string `json:"signature" form:"signature" validate:"required"`
}

// InnerQueryMsgsReq 查询会话内容
type InnerQueryMsgsReq struct {
	QueryChatMsgReq
	ExtCorpID string `json:"ext_corp_id" form:"ext_corp_id" validate:"required"`
	Signature string `json:"signature" form:"signature" validate:"required"`
}

// InnerSearchMsgReq  搜索会话请求
type InnerSearchMsgReq struct {
	SearchMsgReq
	ExtCorpID string `json:"ext_corp_id" form:"ext_corp_id" validate:"required"`
	Signature string `json:"signature" form:"signature" validate:"required"`
}
