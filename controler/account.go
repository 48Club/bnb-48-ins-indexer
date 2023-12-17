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
		tickHashs       = mapset.NewSet[string]()
		inss            = []*dao.InscriptionModel{}
		inssMap         = map[string]*dao.AccountWalletModel{}
		tickHashNeedGet = mapset.NewSet[string]()
	)

	for _, v := range res.Wallet {
		inssMap[v.TickHash] = v
		tickHashs.Add(v.TickHash)
	}

	if tickHashs.Cardinality() != len(req.TickHash) {
		for _, v := range req.TickHash {
			if tickHashs.Contains(v) {
				continue
			}
			tickHashNeedGet.Add(v)
		}

		c.accountS.GetInscription(req.TickHash, &inss)
		for _, v := range inss {
			if tickHashNeedGet.Contains(v.TickHash) {
				inssMap[v.TickHash] = &dao.AccountWalletModel{
					Address:  req.Address,
					Tick:     v.Tick,
					TickHash: v.TickHash,
					Balance:  "0",
					Decimals: v.Decimals,
				}
			}
			if tickHashs.Contains(v.TickHash) {
				_tmp := inssMap[v.TickHash]
				_tmp.Decimals = v.Decimals
				inssMap[v.TickHash] = _tmp
			}

		}
		res.Wallet = []*dao.AccountWalletModel{}
		for _, v := range inssMap {
			res.Wallet = append(res.Wallet, v)
		}
	}

	return nil
}
