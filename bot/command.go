package main

import (
	"fmt"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/ethereum/go-ethereum/common"
)

func commandStart(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.Chat.Id > 0 {
		_, err := ctx.EffectiveMessage.Reply(b, "请直接发送 BEP20 钱包地址", nil)
		if err != nil {
			return fmt.Errorf("start command resp error: %w", err)
		}
	}
	return nil
}

func commandSum(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.Chat.Id == adminUserId {
		user, err := b.GetChatMember(club48ChatId, ctx.EffectiveMessage.From.Id, nil)
		if !isChatAdmin(user.MergeChatMember(), err) {
			return nil
		}

		_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("总参与金额: %.8f fans\n有效金额: 200000/200000 fans,\n待退款金额: %.8f fans", float64(totalSum.Uint64())/1e8, float64(totalReturnSum.Uint64())/1e8), nil)
		return err
	}

	return nil

}

func commandExport(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.From.Id == adminUserId {
		// todo: export data to excel
	}
	return nil
}

func messageEcho(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.Chat.Id == adminUserId {
		user, err := b.GetChatMember(club48ChatId, ctx.EffectiveMessage.From.Id, nil)
		if !isMember(user.MergeChatMember(), err) {
			_, err := ctx.EffectiveMessage.Reply(b, "请先加入 48Club 讨论组 @cn_48club", nil)
			if err != nil {
				return fmt.Errorf("echo command resp error: %w", err)
			}
		} else {
			txt := ctx.EffectiveMessage.Text
			if len(txt) == 42 {
				address := common.HexToAddress(txt)
				u, ok := users[address]
				if !ok {
					_, err := ctx.EffectiveMessage.Reply(b, "没有查询到相关交易记录", nil)
					return err
				}

				_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("钱包地址: %s\n有效预购: %.8f bFans([📶 Tx Hash](https://bscscan.com/tx/%s))\n退款数量: %.8f fans", address.Hex(), float64(u.Validated.Uint64())/1e8, u.txHash.Hex(), float64(u.Returned.Uint64())/1e8), &gotgbot.SendMessageOpts{
					ParseMode: gotgbot.ParseModeMarkdown,
					LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
						IsDisabled: true,
					},
				})
				return err
			} else {
				_, err = ctx.EffectiveMessage.Reply(b, "请输入正确的 BEP20 钱包地址", nil)
				return err
			}
		}
	}
	return nil
}

func isMember(m gotgbot.MergedChatMember, err error) bool {
	if err != nil {
		return false
	}
	s := m.GetStatus()
	if s == "administrator" || s == "creator" || s == "member" {
		return true
	}
	if s == "restricted" && m.CanSendMessages {
		return true
	}
	return false
}

func isChatAdmin(m gotgbot.MergedChatMember, err error) bool {
	if err != nil {
		return false
	}
	s := m.GetStatus()
	if s == "administrator" || s == "creator" {
		return true
	}
	return false
}
