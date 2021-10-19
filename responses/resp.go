package responses

import (
	"msg/constants"
)

type ChataDataResp struct {
	// 0表示成功，错误返回非0错误码，需要参看errmsg。Uint32类型
	ErrCode int `json:"errcode"`
	//返回信息，如非空为错误原因。String类型
	ErrMsg   string     `json:"errmsg"`
	ChatData []ChatData `json:"chatdata"` // attention: string like 'chat_data' not meet wx's requirements
}

type ChatData struct {
	//消息的seq值，标识消息的序号。再次拉取需要带上上次回包中最大的seq。Uint64类型，范围0-pow(2, 64)-1
	Seq uint64 `json:"seq"`
	//消息id，消息的唯一标识，企业可以使用此字段进行消息去重。String类型。msgid 以_external结尾的消息，表明该消息是一条外部消息。
	MsgID string `json:"msgid"`
	//加密此条消息使用的公钥版本号。Uint32类型
	PublicKeyVer uint32 `json:"publickey_ver"`
	//使用publickey_ver指定版本的公钥进行非对称加密后base64加密的内容，需要业务方先base64 decode处理后，再使用指定版本的私钥进行解密，得出内容。String类型
	EncryptRandomKey string `json:"encrypt_random_key"`
	//消息密文。需要业务方使用将encrypt_random_key解密得到的内容，与encrypt_chat_msg，传入sdk接口DecryptData, 得到消息明文。String类型
	EncryptChatMsg string `json:"encrypt_chat_msg"`
}

type TextMsg struct {
	MsgCommonField
	Text struct {
		//消息内容。String类型
		Content string `json:"content"`
	} `json:"text"`
}
type MsgCommonField struct {
	//消息id，消息的唯一标识，企业可以使用此字段进行消息去重。String类型
	MsgID string `json:"msgid"`
	//消息动作，目前有send(发送消息)/recall(撤回消息)/switch(切换企业日志)三种类型。String类型
	Action string `json:"action"`
	//消息发送方id。同一企业内容为userid，非相同企业为external_userid。消息如果是机器人发出，也为external_userid。String类型
	From string `json:"from"`
	//消息接收方列表，可能是多个，同一个企业内容为userid，非相同企业为external_userid。数组，内容为string类型
	ToList []string `json:"tolist"`
	//群聊消息的群id。如果是单聊则为空。String类型
	RoomID string `json:"roomid"`
	//消息发送时间戳，utc时间，ms单位。
	MsgTime int64 `json:"msgtime"`
	//文本消息为：text。String类型
	MsgType string `json:"msgtype"`
}

type ImageMsg struct {
	MsgCommonField
	Image struct {
		//图片资源的md5值，供进行校验。String类型
		Md5Sum string `json:"md5sum"`
		// 	图片资源的文件大小。Uint32类型
		Filesize int `json:"filesize"`
		// 媒体资源的id信息。String类型
		Sdkfileid string `json:"sdkfileid"`
	} `json:"image"`
}

type ReVokeMsg struct {
	MsgCommonField
	Revoke struct {
		PreMsgid string `json:"pre_msgid"`
	} `json:"revoke"`
}

type File struct {
	Md5Sum   string `json:"md5sum"`
	Filename string `json:"filename"`
	Fileext  string `json:"fileext"`
	Filesize int    `json:"filesize"`
	//媒体资源的id信息
	Sdkfileid string `json:"sdkfileid"`
}

type ChatMsgResp struct {
	//消息id，消息的唯一标识，企业可以使用此字段进行消息去重。String类型
	MsgID string `gorm:"type:char(128);unique" json:"msgid"`
	//消息动作，目前有send(发送消息)/recall(撤回消息)/switch(切换企业日志)三种类型。String类型
	Action string `gorm:"type:char(8)" json:"action"`
	//消息发送方id。同一企业内容为userid，非相同企业为external_userid。消息如果是机器人发出，也为external_userid。String类型
	From string `gorm:"type:char(32)" json:"from"`
	//消息接收方列表，可能是多个，同一个企业内容为userid，非相同企业为external_userid。数组，内容为string类型
	ToList constants.StringArrayField `gorm:"type:json" json:"tolist"`
	//群聊消息的群id。如果是单聊则为空。String类型
	RoomID string `gorm:"type:char(128)" json:"roomid"`
	//消息发送时间戳，utc时间，ms单位。
	MsgTime int64 `gorm:"type:bigint(64)" json:"msgtime"`
	//文本消息为：text。String类型
	MsgType string `gorm:"type:varchar(32)" json:"msgtype"`
	// 聊天的文本内容
	ContentText string `gorm:"class:FULLTEXT,option:WITH PARSER ngram" json:"content_text"`
	//消息的seq值，标识消息的序号。再次拉取需要带上上次回包中最大的seq。Uint64类型，范围0-pow(2,64)-1
	Seq uint64 `gorm:"type:bigint unsigned" json:"seq"`
	// 发送-接收双方ID hash得到
	SessionID string `json:"session_id"`
	// 会话类型
	SessionType string `json:"session_type"`

	Text struct {
		Content string `json:"content"`
	} `json:"text"`

	Image struct {
		Md5sum    string `json:"md5sum"`
		Sdkfileid string `json:"sdkfileid"`
		Filesize  uint32 `json:"filesize"`
	} `json:"image"`

	File struct {
		FileExt   string `json:"fileext"` //文件类型后缀。String类型
		Md5Sum    string `json:"md5sum"`
		Filename  string `json:"filename"`
		Filesize  int    `json:"filesize"`
		SdkFileid string `json:"sdkfileid"`
	} `json:"file"`

	Revoke struct {
		PreMsgid string `json:"pre_msgid"` // 标识撤回的原消息的msgid。String类型
	} `json:"revoke"`

	Agree struct {
		Userid    string `json:"userid"`     // 同意/不同意协议者的userid，外部企业默认为external_userid。String类型
		AgreeTime string `json:"agree_time"` // 同意/不同意协议的时间，utc时间，ms单位。
	} `json:"agree"`

	Voice struct {
		Md5sum     string `json:"md5sum"`
		VoiceSize  uint32 `json:"voice_size"`
		PlayLength uint32 `json:"play_length"`
		Sdkfileid  string `json:"sdkfileid"`
	} `json:"voice"`

	TODO struct {
	} `json:"todo"`

	Redpacket struct {
		//红包消息类型。1 普通红包、2 拼手气群红包、3 激励群红包。Uint32类型
		uint32 `json:""`
		//wish	红包祝福语。String类型
		Wish string `json:"wish"`
		//totalcnt	红包总个数。Uint32类型
		Totalcnt uint32 `json:"totalcnt"`
		//totalamount	红包总金额。Uint32类型，单位为分。
		Totalamount uint32 `json:"redpacket"`
	} `json:"redpacket"`

	Emotion struct {
		int       `json:""`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Imagesize int    `json:"imagesize"`
		Md5Sum    string `json:"md5sum"`
		Sdkfileid string `json:"sdkfileid"`
	} `json:"emotion"`

	Location struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Address   string  `json:"address"`
		Title     string  `json:"title"`
		Zoom      int     `json:"zoom"`
	} `json:"location"`

	//名片
	Card struct {
		CorpName string `json:"corpname"` //名片所有者所在的公司名称
		UserID   string `json:"userid"`   //名片所有者的id，同一公司是userid，不同公司是external_userid。String类型
	} `json:"card"`

	DocMsg struct {
		Msg        string `json:"msg"`         //类型, 标识在线文档消息类型
		Title      string `json:"title"`       //       在线文档名称
		LinkURL    string `json:"link_url"`    //    在线文档链接
		DocCreator string `json:"doc_creator"` // 在线文档创建者。本企业成员创建为userid；外部企业成员创建为external_userid
	} `json:"doc_msg"`

	Meeting struct {
		MsgID   string `json:"msgid"`   //	msgid 消息id，消息的唯一标识，企业可以使用此字段进行消息去重。String类型
		Action  string `json:"action"`  //	action 消息动作，目前有send(发送消息)/recall(撤回消息)/switch (切换企业日志)三种类型。String类型
		From    string `json:"from"`    //	from   消息发送方id。同一企业内容为userid，非相同企业为external_userid。消息如果是机器人发出，也为external_userid。String类型
		Tolist  string `json:"tolist"`  //  消息接收方列表，可能是多个，同一个企业内容为userid，非相同企业为external_userid。数组，内容为string类型
		Msgtime string `json:"msgtime"` //  消息发送时间戳，utc时间, 单位毫秒。
		Msg     string `json:"msg"`     //  meeting_voice_call。String类型, 标识音频存档消息类型
		Voiceid string `json:"voiceid"` // String类型, 音频id
		//Meeting_voice_call
		// 音频消息内容。包括结束时间、fileid，可能包括多个demofiledata、sharescreendata消息，demofiledata表示文档共享信息，sharescreendata表示屏幕共享信息。Object类型
		//Endtime    音频结束时间。uint32类型
		//Sdkfileid    sdkfileid。音频媒体下载的id。String类型
		//Demofiledata    文档分享对象，Object类型
		//Filename    文档共享名称。String类型
		//Demooperator    文档共享操作用户的id。String类型
		//Starttime    文档共享开始时间。Uint32类型
		//Endtime    文档共享结束时间。Uint32类型
		//Sharescreendata    屏幕共享对象，Object类型
		//Share    屏幕共享用户的id。String类型
		//Starttime    屏幕共享开始时间。Uint32类型
		//Endtime    屏幕共享结束时间。Uint32类型
	} `json:"meeting"`

	Link struct {
		Title       string `json:"title"`       // 消息标题。String类型
		Description string `json:"description"` // 消息描述。String类型
		LinkUrl     string `json:"link_url"`    // 链接url地址。String类型
		ImageUrl    string `json:"image_url"`   //  链接图片url。String类型
	} `json:"link"`
}
