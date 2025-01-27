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

package dbutils

import (
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	leveldb_errors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"nocalhost/pkg/nhctl/log"
	"strconv"
	"syscall"
	"time"
)

// Initial a level db
// If leveldb already exists, return error
func CreateLevelDB(path string, errorIfExist bool) error {
	db, err := leveldb.OpenFile(path, &opt.Options{
		ErrorIfExist: errorIfExist,
	})
	if db != nil {
		_ = db.Close()
	}
	return errors.Wrap(err, "")
}

// Open a leveldb
// If leveldb is corrupted, try to recover it
// If leveldb is EAGAIN, retry to open it in 1 minutes
// If leveldb is missing, return a error instead create one
func OpenLevelDB(path string, readonly bool) (*LevelDBUtils, error) {
	var o *opt.Options
	o = &opt.Options{
		ErrorIfMissing: true,
	}
	if readonly {
		o.ReadOnly = true
	}
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		if leveldb_errors.IsCorrupted(err) {
			log.Log("Recovering leveldb file...")
			db, err = leveldb.RecoverFile(path, nil)
		} else if errors.Is(err, syscall.EAGAIN) {
			for i := 0; i < 300; i++ {
				log.Logf("Another process is accessing leveldb: %s, wait for 0.2s to retry", path)
				time.Sleep(200 * time.Millisecond)
				db, err = leveldb.OpenFile(path, nil)
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			return nil, errors.Wrap(err, "Retry opening leveldb failed")
		}
	}

	dbUtils := &LevelDBUtils{
		readonly: readonly,
		db:       db,
	}

	if !readonly {
		v, err := db.GetProperty("leveldb.num-files-at-level0")
		if err != nil {
			log.LogE(err)
		} else {
			num, err := strconv.Atoi(v)
			if err != nil {
				log.LogE(err)
			}
			if num > 10 {
				log.Logf("Compacting %s", path)
				if err = dbUtils.CompactFirstKey(); err != nil {
					log.LogE(err)
				}
			}
		}
	}

	return dbUtils, nil
}
