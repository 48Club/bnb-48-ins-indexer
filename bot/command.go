package main

import (
	"bytes"
	"fmt"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/ethereum/go-ethereum/common"
	excelize "github.com/xuri/excelize/v2"
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
	if ctx.EffectiveMessage.Chat.Id == club48ChatId {
		user, err := b.GetChatMember(club48ChatId, ctx.EffectiveMessage.From.Id, nil)
		if !isChatAdmin(user.MergeChatMember(), err) {
			return nil
		}

		_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("总参与金额: %.8f fans\n有效映射金额: 200000/200000 fans\n待退款金额: %.8f fans", float64(totalSum.Uint64())/1e8, float64(totalReturnSum.Uint64())/1e8), nil)
		return err
	}

	return nil
}

func commandExport(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.From.Id == adminUserId {
		// todo: export data to excel
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		validatedList := ""
		returnedList := ""
		tableHeader := []interface{}{"钱包地址", "有效映射", "退款数量", "有效交易"}
		table := [][]interface{}{tableHeader}
		for add, user := range users {
			txHash := user.txHash.Hex()
			if user.Validated.Uint64() == 0 {
				txHash = ""
			}
			table = append(table, []interface{}{add.Hex(), fmt.Sprintf("%.8f", float64(user.Validated.Uint64())/1e8), fmt.Sprintf("%.8f", float64(user.Returned.Uint64())/1e8), txHash})
			if user.Validated.Uint64() > 0 {
				validatedList += fmt.Sprintf("%s,%.8f\n", add.Hex(), float64(user.Validated.Uint64())/1e8)
			}
			if user.Returned.Uint64() > 0 {
				returnedList += fmt.Sprintf("%s,%.8f\n", add.Hex(), float64(user.Returned.Uint64())/1e8)
			}
		}
		_ = f.SetSheetName("Sheet1", "bFans")
		for idx, row := range table {
			cell, err := excelize.CoordinatesToCellName(1, idx+1)
			if err != nil {
				return err
			}
			_ = f.SetSheetRow("bFans", cell, &row)
		}

		bf, _ := f.WriteToBuffer()
		_, err := b.SendDocument(ctx.EffectiveMessage.Chat.Id, gotgbot.NamedFile{
			FileName: "bFans.xlsx",
			File:     bf,
		}, nil)
		if err != nil {
			return err
		}
		// 将 validatedList 和 returnedList 写入 io.Reader
		_, err = b.SendDocument(adminUserId, gotgbot.NamedFile{
			FileName: "validated.csv",
			File:     bytes.NewBufferString(validatedList),
		}, nil)
		if err != nil {
			return err
		}
		_, err = b.SendDocument(adminUserId, gotgbot.NamedFile{
			FileName: "returned.csv",
			File:     bytes.NewBufferString(returnedList),
		}, nil)
		return err
	}
	return nil
}

func messageEcho(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveMessage.Chat.Id > 0 {
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
				txHash := fmt.Sprintf("([📶 Tx Hash](https://bscscan.com/tx/%s))", u.txHash.Hex())
				if u.Validated.Uint64() == 0 {
					txHash = ""
				}

				_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("钱包地址: %s\n有效映射: %.8f bFans%s\n退款数量: %.8f fans", address.Hex(), float64(u.Validated.Uint64())/1e8, txHash, float64(u.Returned.Uint64())/1e8), &gotgbot.SendMessageOpts{
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
