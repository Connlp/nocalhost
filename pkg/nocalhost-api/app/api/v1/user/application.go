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

package user

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"nocalhost/internal/nocalhost-api/model"
	"nocalhost/internal/nocalhost-api/service"
	"nocalhost/pkg/nhctl/log"
	"nocalhost/pkg/nocalhost-api/app/api"
	"nocalhost/pkg/nocalhost-api/pkg/errno"
)

// List all authorized user in application
func ListByApplication(c *gin.Context) {
	// userId, _ := c.Get("userId")
	users, err := listByApplication(c, true)
	if err != nil {
		api.SendResponse(c, err, nil)
	}

	api.SendResponse(c, nil, users)
}

// List all unauthorized user in application
func ListNotInApplication(c *gin.Context) {
	// userId, _ := c.Get("userId")

	users, err := listByApplication(c, false)
	if err != nil {
		api.SendResponse(c, nil, users)
	}

	api.SendResponse(c, nil, users)
}

// list user by application
// in application means user has the permission to this application
func listByApplication(c *gin.Context, inApp bool) ([]*model.UserList, error) {
	applicationId := cast.ToUint64(c.Param("id"))
	applicationUsers, err := service.Svc.ApplicationUser().ListByApplicationId(c, applicationId)

	if err != nil {
		log.Error(err)
		return nil, errno.ErrListApplicationUser
	}

	userList, err := service.Svc.UserSvc().GetUserList(c)
	if err != nil {
		log.Error(err)
		return nil, errno.ErrListApplicationUser
	}

	// first list all user
	// then while applicationUsers contain that user
	// put into inApp list

	set := map[uint64]interface{}{}
	for _, au := range applicationUsers {
		set[au.UserId] = "-"
	}

	result := []*model.UserList{}
	for _, user := range userList {
		_, ok := set[user.ID]

		if inApp && ok {
			result = append(result, user)
		} else if !inApp && !ok {
			result = append(result, user)
		}
	}

	return result, nil
}
