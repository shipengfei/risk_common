package ydmsg

import (
	"context"

	"gitlab.miliantech.com/go/client-go/proto/message"
	"gitlab.miliantech.com/infrastructure/log"
	"gitlab.miliantech.com/risk/base/risk_common/utils"
	"go.uber.org/zap"
)

/*
	msgBase := &message.MsgBase{
		Preview: ctt,
		MsgUnion: &message.MsgBase_Text{
			Text: &msgComm.Text{Id: 74260, Content: ctt},
		},
	}
*/
func SendSystemMsg(ctx context.Context, client message.MsgServerClient, memId, targetId int64, msgBase *message.MsgBase) {
	defer utils.SimpleRecover(ctx)

	// 获取会话信息
	in := &message.GetChatInfoByTargetIdRq{UserId: memId, TargetList: []int64{targetId}}
	chatInfo, errChatInfo := client.GetChatInfoByTargetId(ctx, in)
	if errChatInfo != nil {
		log.YError(ctx, "risk_common.sendSystemMsg.GetChatInfoByTargetId", targetId, 0, zap.Error(errChatInfo))
		return
	}

	// 获得chatId
	var chatId int64
	if len(chatInfo.GetInfoList()) == 0 {
		chatInfoCreate, errCreate := client.CreateChat(ctx, &message.CreateChatRq{
			UserId: targetId, TargetId: memId, Type: message.ChatType_System_msg})
		if errCreate != nil {
			log.YError(ctx, "risk_common.sendSystemMsg.CreateChat", targetId, 0, zap.Error(errChatInfo))
			return
		}
		chatId = chatInfoCreate.GetChatInfo().GetChatId()
	} else {
		chatId = chatInfo.GetInfoList()[0].GetChatId()
	}

	if chatId == 0 {
		log.YError(ctx, "risk_common.chatId.required", targetId, 0)
		return
	}

	msgReq := &message.CreateMsgByInternalSystemRq{
		UserId:   memId,
		ChatId:   chatId,
		MsgType:  message.MsgType_Text,
		PushMode: message.MsgPushMode_ONLY_IM_ALL,
		MsgBase:  msgBase,
	}
	res, err := client.CreateMsgByInternalSystem(ctx, msgReq)
	if err != nil {
		log.YError(ctx, "risk_common.createMsgByInternalSystem", targetId, 0, zap.Error(err))
		return
	}
	log.YInfo(ctx, "risk_common.SendSystemMsg.success", targetId, 0,
		zap.Any("go_text", map[string]any{"request": msgReq, "result": res}))
}

func SendSystemMsgWithChatId(ctx context.Context, client message.MsgServerClient, memId, targetId, chatId int64, msgBase *message.MsgBase) {
	defer utils.SimpleRecover(ctx)

	msgReq := &message.CreateMsgByInternalSystemRq{
		UserId:   memId,
		TargetId: targetId,
		ChatId:   chatId,
		MsgType:  message.MsgType_Text,
		PushMode: message.MsgPushMode_ONLY_IM_ALL,
		MsgBase:  msgBase,
	}
	res, err := client.CreateMsgByInternalSystem(ctx, msgReq)
	if err != nil {
		log.YError(ctx, "risk_common.createMsgByInternalSystem", targetId, 0, zap.Error(err))
		return
	}
	log.YInfo(ctx, "risk_common.SendSystemMsg.success", targetId, 0,
		zap.Any("go_text", map[string]any{"request": msgReq, "result": res}))
}

func SendHitMsgWithChatId(ctx context.Context, client message.MsgServerClient, memId, targetId, chatId int64, msgBase *message.MsgBase) {
	defer utils.SimpleRecover(ctx)

	if client == nil {
		return
	}

	msgReq := &message.CreateMsgByInternalSystemRq{
		UserId:        memId,
		TargetId:      targetId,
		ChatId:        chatId,
		MsgType:       message.MsgType_Hint2,
		PushMode:      message.MsgPushMode_ONLY_IM_ALL,
		MsgBase:       msgBase,
		SysMsgSource:  message.SourceOfSystemMsg_CommonSystem,
		NotSysMsgFlag: false,
	}
	res, err := client.CreateMsgByInternalSystem(ctx, msgReq)
	if err != nil {
		log.YError(ctx, "risk_common.SendHitMsgWithChatId", targetId, 0, zap.Error(err))
		return
	}
	log.YInfo(ctx, "risk_common.SendHitMsgWithChatId.success", targetId, 0,
		zap.Any("go_text", map[string]any{"request": msgReq, "result": res}))
}
