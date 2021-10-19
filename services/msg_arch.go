package services

/*
#cgo CFLAGS: -I ./ -I./lib -I../lib
#cgo CXXFLAGS: -I./
#cgo LDFLAGS:  -L../lib  -lWeWorkFinanceSdk_C -ldl

#include "../lib/WeWorkFinanceSdk_C.h"
#include <stdlib.h>
*/
import "C"
import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"hash"
	"msg/common/id_generator"
	"msg/common/log"
	"msg/common/storage"
	"msg/conf"
	"msg/constants"
	"msg/models"
	"msg/requests"
	"msg/responses"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unsafe"
)

type MsgArch struct {
	model models.ChatMsg
}

func NewMsgArch() *MsgArch {
	return &MsgArch{model: models.ChatMsg{}}
}

func newSDK(extCorpID, secret string) (*C.WeWorkFinanceSdk_t, error) {
	corpID := C.CString(extCorpID)
	defer C.free(unsafe.Pointer(corpID))

	secretKey := C.CString(secret)
	defer C.free(unsafe.Pointer(secretKey))

	sdk := C.NewSdk()
	ret := C.Init(sdk, corpID, secretKey)
	if ret != 0 {
		fmt.Printf("init sdk err ret:%d\n", ret)
		return nil, errors.New("init sdk failed")
	}
	return sdk, nil
}

func readPriKey(keyFilepath string) (*rsa.PrivateKey, error) {
	file, err := os.Open(keyFilepath)
	if err != nil {
		log.Sugar.Error("open priKey failed", err)
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		log.Sugar.Error("stat the priKey failed", err)
		return nil, err
	}
	b := make([]byte, stat.Size())
	_, err = file.Read(b)
	if err != nil {
		log.Sugar.Error("read the priKey failed", err)
		return nil, err
	}

	p, _ := pem.Decode(b)
	key, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		log.Sugar.Error("parse the priKey failed", err)
		return nil, err
	}
	return key, nil
}

func fetch(sdk *C.WeWorkFinanceSdk_t, beginAt uint64, batchSize uint32, timeout int, proxy, proxyPwd string) (*C.Slice_t, error) {
	chatDataSlice := C.NewSlice()
	ret := C.GetChatData(sdk, C.ulonglong(beginAt), C.uint(batchSize), C.CString(proxy), C.CString(proxyPwd), C.int(timeout), chatDataSlice)
	if ret != 0 {
		log.Sugar.Error("fetch data failed", ret)
		return nil, errors.New("get chat data failed")
	}
	return chatDataSlice, nil
}

// 需要下载附件的聊天记录类型
var msgContentNeedDownload map[string]struct{}

// Init
// todo 补齐类型
func (o *MsgArch) Init() {
	msgContentNeedDownload = map[string]struct{}{
		constants.MsgTypeImage: {},
	}
}

func (o *MsgArch) Sync(extCorpID string) error {
	sdk, err := newSDK(conf.Settings.WeWork.ExtCorpID, conf.Settings.WeWork.ContactSecret)
	if err != nil {
		err = errors.WithStack(err)
		return err
	}
	return o.fetchAndStore(sdk, conf.Settings.WeWork.PriKeyPath, extCorpID)
}

// todo
// 	1. to_list 多个值的情况还没遇到
func (o *MsgArch) fetchAndStore(sdk *C.WeWorkFinanceSdk_t, priKeyPath string, extCorpID string) (err error) {
	key, err := readPriKey(priKeyPath)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	batchSize := conf.Settings.WeWork.MsgArchBatchSize
	if batchSize == 0 {
		batchSize = 10
	}
	timeout := conf.Settings.WeWork.MsgArchTimeout
	if timeout == 0 {
		timeout = 10 // second
	}

	h := md5.New()
	noMoreMsg := false
	for !noMoreMsg {
		seq, err := o.model.GetLatestSeq(extCorpID)
		if err != nil {
			err = errors.WithStack(err)
			return err
		}

		chatDataSlice, err := fetch(sdk,
			uint64(seq+1),
			uint32(batchSize),
			conf.Settings.WeWork.MsgArchTimeout,
			conf.Settings.WeWork.MsgArchProxy,
			conf.Settings.WeWork.MsgArchProxyPasswd,
		)
		if err != nil {
			log.Sugar.Errorw("fetch msg data failed", "err", err)
			err = errors.WithStack(err)
			return err
		}
		chatData := responses.ChataDataResp{}
		err = json.Unmarshal([]byte(C.GoString(chatDataSlice.buf)), &chatData)
		if err != nil {
			log.Sugar.Error("unmarshal chat_data failed", "err", err)
			return err
		}
		if chatData.ErrCode != 0 {
			log.Sugar.Errorf("msg arch responses code:%d", chatData.ErrCode)
			return err
		}

		chatMsgs := make([]models.ChatMsg, 0)
		if len(chatData.ChatData) == 0 {
			noMoreMsg = true
		} else {
			for _, item := range chatData.ChatData {
				bytes, err := base64.StdEncoding.DecodeString(item.EncryptRandomKey)
				if err != nil {
					log.Sugar.Error("DecodeString failed", err)
					continue
				}

				v15, err := rsa.DecryptPKCS1v15(rand.Reader, key, bytes)
				if err != nil {
					log.Sugar.Errorw("decrypt random key failed", "err", err, "content", item.EncryptRandomKey)
					continue
				}

				decryptSlice := C.NewSlice()
				ret := C.DecryptData(C.CString(string(v15)), C.CString(item.EncryptChatMsg), decryptSlice)
				if ret != 0 {
					log.Sugar.Error("DecryptData in so failed,ret:", ret)
					continue
				}

				msgStr := C.GoString(decryptSlice.buf)
				//log.Sugar.Debug("msgResp", msgStr)

				resp := responses.ChatMsgResp{}
				err = json.Unmarshal([]byte(msgStr), &resp)
				if err != nil {
					log.Sugar.Errorw("unmarshal the chat msg responses to model failed", "err", err)
					continue
				}

				msg := models.ChatMsg{
					ExtCorpModel: models.ExtCorpModel{
						ID: id_generator.StringID(), ExtCorpID: conf.Settings.WeWork.ExtCorpID, ExtCreatorID: resp.From,
					},
				}
				err = copier.CopyWithOption(&msg, resp, copier.Option{
					IgnoreEmpty: true,
					DeepCopy:    true,
				})
				if err != nil {
					err = errors.WithStack(err)
					return err
				}
				msg.Seq = item.Seq

				// 1. 用字符串 peerA+peerB 生成会话hash, peerA/peerB 按字典序排序
				// 2. 判断会话类型 internal/external/room
				h.Reset()
				msg.SessionID = o.getSessionID(msg, h)
				msg.SessionType = o.getSessionType(msg)
				msg.Keywords = o.getSessionKeywords(msg)

				if msg.MsgType == constants.MsgTypeTextMsg {
					msg.ContentText = resp.Text.Content
				}

				// 下载有附件的消息,写到关联表ChatMsgContent
				if _, ok := msgContentNeedDownload[resp.MsgType]; ok {

					mediaData, obj, err := o.downloadFromWx(sdk, resp)
					if err != nil {
						err = errors.Wrap(err, "download chat msg from wx failed")
						return err
					}

					err = storage.FileStorage.Put(obj, strings.NewReader(C.GoString(mediaData.data)))
					if err != nil {
						err = errors.Wrap(err, "upload chat msg file to storage failed")
						log.Sugar.Errorw("upload failed", "err", err)
						return err
					}

					downloadURL, err := o.createFileDownloadURL(resp, http.MethodGet)
					if err != nil {
						err = errors.Wrap(err, "createFileDownloadURL failed")
						return err
					}
					content := models.ChatMsgContent{
						ExtCorpModel: models.ExtCorpModel{
							ID:           id_generator.StringID(),
							ExtCorpID:    conf.Settings.WeWork.ExtCorpID,
							ExtCreatorID: msg.From,
						},
						ContentType: msg.MsgType,
						FileURL:     downloadURL,
					}
					contentJsonStr, err := o.createMsgContentJsonStr(resp)
					if err != nil {
						err = errors.Wrap(err, "createMsgContentJsonStr failed")
						return err
					}
					content.Content = string(contentJsonStr)
					msg.ChatMsgContent = content
				}
				chatMsgs = append(chatMsgs, msg)
			}
			err = o.model.Creat(chatMsgs)
			if err != nil {
				err = errors.WithStack(err)
				return err
			}
		}
	}
	return err
}

func (o *MsgArch) getSessionID(msg models.ChatMsg, h hash.Hash) string {
	if msg.RoomID != "" {
		h.Write([]byte(msg.From + msg.RoomID))
	} else {
		if msg.From < msg.ToList[0] {
			h.Write([]byte(msg.From + msg.ToList[0]))
		} else {
			h.Write([]byte(msg.ToList[0] + msg.From))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (o *MsgArch) getSessionType(msg models.ChatMsg) (sessionType string) {
	if msg.RoomID != "" {
		sessionType = string(constants.ChatSessionTypeGroup)
	} else {
		isExternal := strings.HasSuffix(msg.MsgID, string(constants.ChatSessionTypeExternal))
		if isExternal {
			sessionType = string(constants.ChatSessionTypeExternal)
		} else {
			sessionType = string(constants.ChatSessionTypeInternal)
		}
	}
	return
}

func (o *MsgArch) downloadFromWx(
	sdk *C.WeWorkFinanceSdk_t, resp responses.ChatMsgResp) (mediaData *C.MediaData_t, obj string, err error) {

	var fileid string
	mediaData = C.NewMediaData()
	switch resp.MsgType {
	case constants.MsgTypeImage:
		fileid = resp.Image.Sdkfileid
	case constants.MsgTypeFile:
		fileid = resp.File.FileExt

	default:
	}
	ret := C.GetMediaData(
		sdk, C.CString(""),
		C.CString(fileid),
		C.CString(conf.Settings.WeWork.MsgArchProxy),
		C.CString(conf.Settings.WeWork.MsgArchProxyPasswd),
		C.int(conf.Settings.WeWork.MsgArchTimeout), mediaData)
	if ret != 0 {
		fmt.Printf("get media data failed, ret:%d\n", ret)
		err = errors.New("get media data failed")
		return
	}

	//文件命名 extcorpid/msgarch/msg_id.文件后缀
	var ext string
	if resp.MsgType == constants.MsgTypeFile {
		ext = resp.File.FileExt
	} else if resp.MsgType == constants.MsgTypeImage {
		ext = "jpg"
	}
	obj = path.Join(conf.Settings.WeWork.ExtCorpID, "/", constants.ModuleNameMsgArch, "/", resp.MsgID+"."+ext)
	return
}

func (o MsgArch) createFileDownloadURL(resp responses.ChatMsgResp, method constants.HTTPMethod) (url string, err error) {
	var ext string

	if resp.MsgType == constants.MsgTypeFile {
		ext = resp.File.FileExt
	} else if resp.MsgType == constants.MsgTypeImage {
		ext = "jpg"
	}
	obj := path.Join(conf.Settings.WeWork.ExtCorpID, "/", constants.ModuleNameMsgArch, "/", resp.MsgID+"."+ext)

	return storage.FileStorage.SignURL(obj, method, 365*24*3600)
}

// todo 补齐类型
func (o MsgArch) createMsgContentJsonStr(resp responses.ChatMsgResp) (contentStr []byte, err error) {
	switch resp.MsgType {
	case constants.MsgTypeTextMsg:
		return json.Marshal(resp.Text)
	case constants.MsgTypeImage:
		return json.Marshal(resp.Image)
	case constants.MsgTypeFile:
		return json.Marshal(resp.File)
	case constants.MsgTypeLink:
		return json.Marshal(resp.Link)
	case constants.MsgTypeDocMsg:
		return json.Marshal(resp.DocMsg)
	case constants.MsgTypeVoice:
		return json.Marshal(resp.Voice)
	case constants.MsgTypeCard:
		return json.Marshal(resp.Card)
	default:
		err = errors.New("uncompleted type")
	}
	return
}

func (o MsgArch) QuerySessions(
	req requests.QuerySessionReq, extCorpID string) (resp []models.ChatSessions, total int64, err error) {

	resp, total, err = o.model.QuerySessions(req.ExtStaffID, req.SessionType, extCorpID, req.Name, &req.Pager)
	if err != nil {
		log.Sugar.Error("query sessions failed", err)
		err = errors.WithStack(err)
		return
	}

	return
}

func (o MsgArch) QueryMsgs(req requests.QueryChatMsgReq, extCorpID string) (resp []models.ChatMessage, total int64, err error) {

	var startTime time.Time
	var endTime time.Time
	if req.SendAtStart != "" && req.SendAtEnd != "" {
		startTime, err = time.Parse(constants.DateLayout, string(req.SendAtStart))
		if err != nil {
			log.Sugar.Error(err)
			err = errors.WithStack(err)
			return
		}

		endTime, err = time.Parse(constants.DateLayout, string(req.SendAtEnd))
		if err != nil {
			log.Sugar.Error(err)
			err = errors.WithStack(err)
			return
		}
	}

	resp, total, err = o.model.QueryMsg(req, extCorpID, &startTime, &endTime, &req.Sorter, &req.Pager)
	if err != nil {
		log.Sugar.Error(err)
		err = errors.WithStack(err)
		return
	}
	return
}

func (o *MsgArch) SearchMsgs(req requests.SearchMsgReq, extCorpID string) (resp []models.ChatMessage, total int64, err error) {
	msg := models.ChatMsg{
		ExtCorpModel: models.ExtCorpModel{ExtCorpID: extCorpID},
		From:         req.ExtStaffID,
		ToList:       []string{req.ExtPeerID},
		ContentText:  req.Keyword,
	}
	resp, total, err = o.model.SearchMsg(msg, &req.Pager)
	if err != nil {
		log.Sugar.Error(err)
		err = errors.WithStack(err)
		return
	}
	return
}

// getSessionKeywords
// Description: 按客户名搜索的词
func (o *MsgArch) getSessionKeywords(msg models.ChatMsg) (keyword string) {
	switch msg.SessionType {
	case string(constants.ChatSessionTypeInternal):

	case string(constants.ChatSessionTypeExternal):
	case string(constants.ChatSessionTypeGroup):
	default:

	}
	return
}
