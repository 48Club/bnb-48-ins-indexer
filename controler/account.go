package controler

import (
	"bnb-48-ins-indexer/dao"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"

	mapset "github.com/deckarep/golang-set/v2"
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
		tickHashs    = mapset.NewSet[string]()
		inss         = []*dao.InscriptionModel{}
		insTickHashs = []string{}
	)
	for _, v := range req.TickHash {
		tickHashs.Add(v)
	}
	if tickHashs.Cardinality() != len(req.TickHash) {
		for _, v := range req.TickHash {
			if tickHashs.Contains(v) {
				continue
			}
			insTickHashs = append(insTickHashs, v)
		}
		c.accountS.GetInscription(insTickHashs, &inss)
		for _, v := range inss {
			res.Wallet = append(res.Wallet, &dao.AccountWalletModel{
				Address:  req.Address,
				Tick:     v.Tick,
				TickHash: v.TickHash,
				Balance:  "0",
				Decimals: v.Decimals,
			})
		}

	}

	return nil
}
