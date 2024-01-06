package helper

type AccountRecordsType uint8

const (
	InscriptionStatusUndefined AccountRecordsType = iota
	InscriptionStatusMint
	InscriptionStatusTransfer
	InscriptionStatusRecap
	InscriptionStatusBurn
	InscriptionStatusApprove
	InscriptionStatusTransferFrom
)
