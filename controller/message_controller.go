package controller

import (
	"server/core/ewc"
	"server/model/dao"
)

type MessageCtrl struct {
	config  *dao.Config
	service *ewc.DbMessageService
}

func NewMessageCtrl(cfg *dao.Config) *MessageCtrl {
	ctrl := new(MessageCtrl)
	ctrl.config = cfg
	ctrl.service = ewc.NewDbMessageService(cfg.Driver, cfg.ConnectionString)

	return ctrl
}
