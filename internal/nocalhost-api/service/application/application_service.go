/*
 * Tencent is pleased to support the open source community by making Nocalhost available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package application

import (
	"context"
	"nocalhost/internal/nocalhost-api/model"
	"nocalhost/internal/nocalhost-api/repository/application"

	"github.com/pkg/errors"
)

type ApplicationService interface {
	Create(ctx context.Context, context string, status uint8, public uint8, userId uint64) (
		model.ApplicationModel, error,
	)
	Get(ctx context.Context, id uint64) (model.ApplicationModel, error)
	GetByName(ctx context.Context, name string) (model.ApplicationModel, error)
	PluginGetList(ctx context.Context, userId uint64) ([]*model.PluginApplicationModel, error)
	GetList(ctx context.Context, userId *uint64) ([]*model.ApplicationModel, error)
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, applicationModel *model.ApplicationModel) (*model.ApplicationModel, error)
	PublicSwitch(ctx context.Context, applicationId uint64, public uint8) error
	Close()
}

type applicationService struct {
	applicationRepo application.ApplicationRepo
}

func NewApplicationService() ApplicationService {
	db := model.GetDB()
	return &applicationService{
		applicationRepo: application.NewClusterRepo(db),
	}
}

func (srv *applicationService) PublicSwitch(ctx context.Context, applicationId uint64, public uint8) error {
	return srv.applicationRepo.PublicSwitch(ctx, applicationId, public)
}

func (srv *applicationService) GetByName(ctx context.Context, name string) (model.ApplicationModel, error) {
	return srv.applicationRepo.GetByName(ctx, name)
}

func (srv *applicationService) PluginGetList(ctx context.Context, userId uint64) (
	[]*model.PluginApplicationModel, error,
) {
	return srv.applicationRepo.PluginGetList(ctx, userId)
}

func (srv *applicationService) Create(
	ctx context.Context, context string, status uint8, public uint8, userId uint64,
) (model.ApplicationModel, error) {
	c := model.ApplicationModel{
		Context: context,
		UserId:  userId,
		Status:  status,
		Public:  public,
	}
	result, err := srv.applicationRepo.Create(ctx, c)
	if err != nil {
		return result, errors.Wrapf(err, "create application")
	}
	return result, nil
}

func (srv *applicationService) Get(ctx context.Context, id uint64) (model.ApplicationModel, error) {
	result, err := srv.applicationRepo.Get(ctx, id)
	if err != nil {
		return result, errors.Wrapf(err, "get application")
	}
	return result, nil
}

func (srv *applicationService) GetList(ctx context.Context, userId *uint64) ([]*model.ApplicationModel, error) {
	result, err := srv.applicationRepo.GetList(ctx, userId)
	if err != nil {
		return nil, errors.Wrapf(err, "get application")
	}
	return result, nil
}

func (srv *applicationService) Delete(ctx context.Context, id uint64) error {
	err := srv.applicationRepo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "delete application error")
	}
	return nil
}

func (srv *applicationService) Update(
	ctx context.Context, applicationModel *model.ApplicationModel,
) (*model.ApplicationModel, error) {
	_, err := srv.applicationRepo.Update(ctx, applicationModel)
	if err != nil {
		return applicationModel, errors.Wrapf(err, "update application error")
	}
	return applicationModel, nil
}

func (srv *applicationService) Close() {
	srv.applicationRepo.Close()
}
