package controler

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/types"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"
	"log"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	accountS   *service.AccountService
	walletDao  dao.IAccountWallet
	pendingTxs *types.GlobalVariable
}

func NewAccountController(pendingTxs *types.GlobalVariable) *AccountController {
	return &AccountController{
		accountS:   service.NewAccountService(),
		walletDao:  &dao.AccountWalletHandler{},
		pendingTxs: pendingTxs,
	}
}

func (c *AccountController) List(ctx *gin.Context) {
	var req bnb48types.ListAccountWalletReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.accountS.List(req, c.pendingTxs.IndexBloukAt)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}

func (c *AccountController) Balance(ctx *gin.Context) {
	db := database.Mysql()

	var req bnb48types.AccountBalanceReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.accountS.Balance(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	for _, v := range res.Wallet {
		if v.CreateAt == 0 {
			v.Address = req.Address
			v.Balance = "0"
			continue
		}
		if _txsByAddr, ok := c.pendingTxs.TxsByAddr[v.Address]; ok {
			if _txsByTickHash, ok := _txsByAddr[v.TickHash]; ok {
				for _, tx := range _txsByTickHash {
					v.Changes = append(v.Changes, *tx)
				}
			}
		}
		if c.walletDao.LoadChanges(db, v, len(v.Changes)) != nil {
			log.Println("LoadChanges err:", err)
		}
	}

	utils.SuccessResponse(ctx, res)
}
