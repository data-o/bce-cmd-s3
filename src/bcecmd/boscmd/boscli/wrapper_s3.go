// Copyright 2017 Baidu, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

// This file provide functions for client.

package boscli

import (
    "fmt"
    "time"
)

import (
	"utils/util"
)

func getObjectMeta(bosClient bosClientInterface, bucketName,
	objectKey string) (*fileDetail, error) {
	if bucketName == "" || objectKey == "" {
		return nil, fmt.Errorf("bucket name and object name can not be empty!")
	}

	getMetaRet, err := bosClient.GetObjectMeta(bucketName, objectKey)
	if err != nil {
		return nil, err
	}
	// utc to timestamp
	mtime, err := util.TranUTCTimeStringToTimeStamp(getMetaRet.LastModified, BOS_HTTP_TIME_FORMT)
	if err != nil {
		return nil, err
	}

	fileInfo := &fileDetail{
		path:  objectKey,
		mtime: mtime,
		gtime: time.Now().Unix(),
		size:  *getMetaRet.ContentLength,
	}
	if getMetaRet.StorageClass != nil {
		fileInfo.storageClass = *getMetaRet.StorageClass
	}
	return fileInfo, nil
}
