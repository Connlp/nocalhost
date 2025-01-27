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

package redis

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Lock
type Lock struct {
	key         string
	redisClient *redis.Client
	timeout     time.Duration
}

// NewLock
func NewLock(conn *redis.Client, key string, defaultTimeout time.Duration) *Lock {
	return &Lock{
		key:         key,
		redisClient: conn,
		timeout:     defaultTimeout,
	}
}

// Lock
func (l *Lock) Lock(token string) (bool, error) {
	ok, err := l.redisClient.SetNX(l.GetKey(), token, l.timeout).Result()
	if err == redis.Nil {
		err = nil
	}
	return ok, err
}

// Unlock
func (l *Lock) Unlock(token string) error {
	script := "if redis.call('get',KEYS[1]) == ARGV[1] then return redis.call('del',KEYS[1]) else return 0 end"
	_, err := l.redisClient.Eval(script, []string{l.GetKey()}, token).Result()
	if err != nil {
		return err
	}
	return nil
}

// GetKey
func (l *Lock) GetKey() string {
	keyPrefix := viper.GetString("name")
	lockKey := "redis:lock"
	return strings.Join([]string{keyPrefix, lockKey, l.key}, ":")
}

// GenToken
func (l *Lock) GenToken() string {
	u, _ := uuid.NewRandom()
	return u.String()
}
