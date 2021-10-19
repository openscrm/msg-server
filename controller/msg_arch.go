package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"msg/common/app"
	"msg/common/ecode"
	"msg/common/log"
	"msg/common/util"
	"msg/conf"
	"msg/requests"
	"msg/responses"
	"msg/services"
)

type MsgArch struct {
	srv *services.MsgArch
}

func CheckMAC(message []byte, messageMAC string, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	log.Sugar.Debugw("Isigned", ">>", hex.EncodeToString(expectedMAC))
	return hex.EncodeToString(expectedMAC) == messageMAC
}

// Sync 同步聊天记录
func (o MsgArch) Sync(c *gin.Context) {
	handler := app.NewDummyHandler(c)
	req := requests.SyncReq{}
	ok, err := handler.BindAndValidateReq(&req)
	if !ok {
		handler.ResponseBadRequestError(errors.WithStack(err))
		return
	}

	if !CheckMAC([]byte(req.ExtCorpID), req.Signature, []byte(conf.Settings.App.InnerSrvAppCode)) {
		err := ecode.CheckSignFailed
		log.TracedError("QuerySessions failed", err)
		handler.ResponseError(err)
		return
	}

	err = o.srv.Sync(req.ExtCorpID)
	if err != nil {
		log.TracedError("QuerySessions failed", err)
		handler.ResponseError(err)
		return
	}

	handler.ResponseItem(nil)
}

// QuerySessions
// 查询会话
func (o MsgArch) QuerySessions(c *gin.Context) {
	req := requests.InnerQuerySessionsReq{}
	if err := app.BindAndValid(c, &req); err != nil {
		app.ResponseErr(c, err)
		return
	}
	needSignBytes, err := util.GenBytesOrderByColumn(req.QuerySessionReq)
	if err != nil {
		app.ResponseErr(c, err)
		return
	}
	log.Sugar.Debugw("req", ">>", req)
	log.Sugar.Debugw("Signature", ">>", req.Signature)
	if !CheckMAC(needSignBytes, req.Signature, []byte(conf.Settings.App.InnerSrvAppCode)) {
		err := ecode.CheckSignFailed
		log.TracedError("CheckSignFailed", err)
		app.ResponseErr(c, err)
		return
	}

	sessionItems, total, err := o.srv.QuerySessions(req.QuerySessionReq, req.ExtCorpID)
	if err != nil {
		log.TracedError("QuerySessions failed", err)
		app.ResponseErr(c, err)
		return
	}

	app.ResponseItems(c, sessionItems, total)
}

// QueryChatMsgs
// 查询某个会话聊天记录
func (o MsgArch) QueryChatMsgs(c *gin.Context) {
	req := requests.InnerQueryMsgsReq{}
	if err := app.BindAndValid(c, &req); err != nil {
		app.ResponseErr(c, errors.WithStack(err))
		return
	}

	bytes, err := util.GenBytesOrderByColumn(req.QueryChatMsgReq)
	if err != nil {
		app.ResponseErr(c, err)
		return
	}
	if !CheckMAC(bytes, req.Signature, []byte(conf.Settings.App.InnerSrvAppCode)) {
		err := ecode.CheckSignFailed
		log.TracedError("Query msgs failed", err)
		app.ResponseErr(c, err)
		return
	}

	Msgs, total, err := o.srv.QueryMsgs(req.QueryChatMsgReq, req.ExtCorpID)
	if err != nil {
		log.TracedError("Query msgs failed", err)
		app.ResponseErr(c, err)
		return
	}

	app.ResponseItem(c, responses.InnerMsgArchSerMsgResp{Items: Msgs, Total: total})
}

// SearchMsgs
// 搜索文字内容
func (o MsgArch) SearchMsgs(c *gin.Context) {
	req := requests.InnerSearchMsgReq{}
	if err := app.BindAndValid(c, &req); err != nil {
		app.ResponseErr(c, errors.WithStack(err))
		return
	}

	bytes, err := util.GenBytesOrderByColumn(req.SearchMsgReq)
	if err != nil {
		app.ResponseErr(c, err)
		return
	}
	if !CheckMAC(bytes, req.Signature, []byte(conf.Settings.App.InnerSrvAppCode)) {
		err = ecode.CheckSignFailed
		log.TracedError("CheckMAC failed", err)
		app.ResponseErr(c, err)
		return
	}

	Msgs, total, err := o.srv.SearchMsgs(req.SearchMsgReq, req.ExtCorpID)
	if err != nil {
		log.TracedError("Search msgs failed", err)
		app.ResponseErr(c, err)
		return
	}

	app.ResponseItem(c, responses.InnerMsgArchSerMsgResp{Items: Msgs, Total: total})

}

func NewMsgArch() *MsgArch {
	return &MsgArch{srv: services.NewMsgArch()}
}
