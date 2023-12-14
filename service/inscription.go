package service

import "github.com/jwrookie/fans/dao"

type InscriptionService struct {
	inscriptionDao dao.IInscription
}

func NewInscriptionService() *InscriptionService {
	return &InscriptionService{
		inscriptionDao: &dao.InscriptionHandler{},
	}
}
