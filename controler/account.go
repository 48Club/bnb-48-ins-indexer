package controler

import (
	"bnb-48-ins-indexer/dao"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	accountS *service.AccountService
}

func NewAccountController() *AccountController {
	return &AccountController{
		accountS: service.NewAccountService(),
	}
}

func (c *AccountController) List(ctx *gin.Context) {
	var req bnb48types.ListAccountWalletReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.accountS.List(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}

func (c *AccountController) Balance(ctx *gin.Context) {
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

	if c.checkBalance(req, res) != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}

func (c *AccountController) checkBalance(req bnb48types.AccountBalanceReq, res *bnb48types.AccountBalanceRsp) error {
	var (
		inss               = []*dao.InscriptionModel{}
		decimalsByTickHash = map[string]uint8{}
	)

	if err := c.accountS.GetInscription(req.TickHash, &inss); err != nil || len(inss) == 0 {
		return err
	}

	for _, v := range inss {
		decimalsByTickHash[v.TickHash] = v.Decimals
	}

	for k, v := range res.Wallet {
		res.Wallet[k].Decimals = decimalsByTickHash[v.TickHash]
	}

	return nil
}
