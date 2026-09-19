package models

import "github.com/digitalwayhk/core/pkg/persistence/entity"

// ServiceModel 是 Pulse 服务的公共持久化基座，统一本地/远程库名。
type ServiceModel struct {
	*entity.Model
}

func newServiceModel() *ServiceModel {
	return &ServiceModel{Model: entity.NewModel()}
}

func (own *ServiceModel) ensureModel() {
	if own.Model == nil {
		own.Model = entity.NewModel()
	}
}

// GetLocalDBName 返回 Pulse 本地 SQLite 库名。
func (*ServiceModel) GetLocalDBName() string { return "pulse" }

// GetRemoteDBName 返回 Pulse 远程库名，开发环境与本地库同名。
func (*ServiceModel) GetRemoteDBName() string { return "pulse" }

// BaseDataModel 是基础资料支路，承载相对稳定、可供引用的配置。
type BaseDataModel struct {
	*ServiceModel
}

func newBaseDataModel() *BaseDataModel {
	return &BaseDataModel{ServiceModel: newServiceModel()}
}

func (own *BaseDataModel) ensureBase() {
	if own.ServiceModel == nil {
		own.ServiceModel = newServiceModel()
		return
	}
	own.ensureModel()
}

// BusinessModel 是业务事实支路，承载持续产生的行情快照与用户操作。
type BusinessModel struct {
	*ServiceModel
}

func newBusinessModel() *BusinessModel {
	return &BusinessModel{ServiceModel: newServiceModel()}
}

func (own *BusinessModel) ensureBusiness() {
	if own.ServiceModel == nil {
		own.ServiceModel = newServiceModel()
		return
	}
	own.ensureModel()
}
