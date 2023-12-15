package types

import "github.com/jwrookie/fans/dao"

type CommonListCond struct {
	Page     int64 `json:"page"`
	PageSize int8  `json:"page_size"  binding:"required"`
}

type CommonListRsp struct {
	Count    int64  `json:"count"`
	Page     uint64 `json:"page"`
	PageSize uint8  `json:"page_size"`
}

type ListAccountWalletReq struct {
	CommonListCond
	TickHash string `json:"tick_hash"  binding:"required"`
}

type ListAccountWalletRsp struct {
	CommonListRsp
	List []*dao.AccountWalletModel `json:"list"`
}

type ListRecordRsp struct {
	CommonListRsp
	List []*dao.AccountRecordsModel `json:"list"`
}

type ListInscriptionWalletReq struct {
	CommonListCond
	Protocol string `json:"protocol"`
	TickHash string `json:"tick_hash"`
	Status   uint64 `json:"status"`
}

type ListInscriptionRsp struct {
	CommonListRsp
	List []*dao.InscriptionModel `json:"list"`
}
