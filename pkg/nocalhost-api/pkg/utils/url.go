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

package utils

import (
	"strings"

	"github.com/qiniu/api.v7/storage"
	"github.com/spf13/viper"

	"nocalhost/pkg/nocalhost-api/pkg/constvar"
)

// GetDefaultAvatarURL
func GetDefaultAvatarURL() string {
	return GetQiNiuPublicAccessURL(constvar.DefaultAvatar)
}

// GetAvatarURL user's avatar, if empty, use default avatar
func GetAvatarURL(key string) string {
	if key == "" {
		return GetDefaultAvatarURL()
	}
	if strings.HasPrefix(key, "https://") {
		return key
	}
	return GetQiNiuPublicAccessURL(key)
}

// GetQiNiuPublicAccessURL
func GetQiNiuPublicAccessURL(path string) string {
	domain := viper.GetString("qiniu.cdn_url")
	key := strings.TrimPrefix(path, "/")

	publicAccessURL := storage.MakePublicURL(domain, key)

	return publicAccessURL
}
