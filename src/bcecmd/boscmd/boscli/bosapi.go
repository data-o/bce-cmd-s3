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

// This module provides the major operations on BOS API.

package boscli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Create new BosApi
func NewBosApi() *BosApi {
	var (
		ak       string
		sk       string
		endpoint string
		err      error
	)

	bosapiClient := &BosApi{}
	bosapiClient.bosClient, err = bosClientInit(ak, sk, endpoint)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
	return bosapiClient
}

type BosApi struct {
	bosClient bosClientInterface
}

type putBucketAclArgs struct {
	bucketName string
	acl        []byte
	opType     int
}

// Put ACL
func (b *BosApi) PutBucketAcl(aclConfigPath, bosPath string, canned string) {

	// preprocessing
	// opType:
	//    1 put acl from file
	//    2 put acl from canned
	args, err, retCode := b.putBucketAclPreProcess(aclConfigPath, bosPath, canned)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}

	//executing
	err, retCode = b.putBucketAclExecute(args.opType, args.acl, args.bucketName, canned)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCodeErr(retCode, err)
	}
}

// Put ACL preprocessing
func (b *BosApi) putBucketAclPreProcess(aclConfigPath, bosPath, canned string) (*putBucketAclArgs,
	error, BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return nil, nil, retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return nil, nil, BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return nil, nil, BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	if aclConfigPath != "" && canned != "" {
		return nil, fmt.Errorf("Can't put acl from canned and file at the same time"),
			BOSCLI_PUT_ACL_CANNED_FILE_SAME_TIME
	}

	// is canned acl?
	if canned != "" {
		if canned != "private" && canned != "public-read" && canned != "public-read-write" {
			return nil, fmt.Errorf("usupported canned ACL"), BOSCLI_PUT_ACL_CANNED_DONT_SUPPORT
		}
		return &putBucketAclArgs{bucketName: bucketName, opType: 2}, nil, BOSCLI_OK
	} else {
		return nil, nil, BOSCLI_PUT_ACL_CANNED_FILE_BOTH_EMPTY
	}
}

// Executing put acl
func (b *BosApi) putBucketAclExecute(opType int, aclJosn []byte, bucketName, canned string) (error,
	BosCliErrorCode) {

	// put canned ACL
	if opType == 2 {
		if err := b.bosClient.PutBucketAclFromCanned(bucketName, canned); err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
	} else {
		// print acl from file.
		var out bytes.Buffer
		json.Indent(&out, aclJosn, "", "  ")
		out.WriteTo(os.Stdout)

		// put ACL
		if err := b.bosClient.PutBucketAclFromString(bucketName, string(aclJosn)); err != nil {
			return err, BOSCLI_EMPTY_CODE
		}
	}
	return nil, BOSCLI_OK
}

// Get ACL
// must have bucket_name
func (b *BosApi) GetBucketAcl(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getBucketAclPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	if err := b.getBucketAclExecute(bucketName); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// Get ACL preprocessing
func (b *BosApi) getBucketAclPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

// Get ACL execute
func (b *BosApi) getBucketAclExecute(bucketName string) error {
	// get Acl
	ret, err := b.bosClient.GetBucketAcl(bucketName)
	if err != nil {
		return err
	}

	fmt.Println(ret.String())

	// print ACL
	/*
		aclJosn, err := json.Marshal(ret)
		if err != nil {
			return err
		}
		var out bytes.Buffer
		json.Indent(&out, aclJosn, "", "  ")
		fmt.Println(out.String())
	*/
	return nil
}

// Put storage class
// must have bucket-name and storage-class
func (b *BosApi) PutBucketStorageClass(bosPath, storageClass string) {
	// check request
	bucketName, retCode := b.putBucketStorageClassPreProcess(bosPath, storageClass)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// put storage class
	if err := b.putBucketStorageClassExecute(bucketName, storageClass); err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// put storage class preprocessing
func (b *BosApi) putBucketStorageClassPreProcess(bosPath, storageClass string) (string,
	BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

	// check storage class is valid
	if storageClass == "" {
		return "", BOSCLI_STORAGE_CLASS_IS_EMPTY
	}
	if _, retCode := getStorageClassFromStr(storageClass); retCode != BOSCLI_OK {
		return "", retCode
	}
	return bucketName, BOSCLI_OK
}

// Put storage class execute
func (b *BosApi) putBucketStorageClassExecute(bucketName, storageClass string) error {
	return b.bosClient.PutBucketStorageclass(bucketName, storageClass)
}

// Get storage class
// must have bucket-name and storage-class
func (b *BosApi) GetBucketStorageClass(bosPath string) {
	// check bucket name
	bucketName, retCode := b.getBucketStorageClassPreProcess(bosPath)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// get storage class
	err := b.getBucketStorageClassExecute(bucketName)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// get storage class preprocessing
func (b *BosApi) getBucketStorageClassPreProcess(bosPath string) (string, BosCliErrorCode) {
	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", retCode
	}
	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}
	return bucketName, BOSCLI_OK
}

func (b *BosApi) getBucketStorageClassExecute(bucketName string) error {
	// get storage class
	ret, err := b.bosClient.GetBucketStorageclass(bucketName)
	if err != nil {
		return err
	}

	// print logging information
	fmt.Printf("{\n    \"storageClass\": \"%s\"\n}\n", ret)

	return nil
}

// for command get-object-meta
func (b *BosApi) HeadObject(bucketName, objectName string) {
// check bucket name
	bucketName, objectName, retCode := b.headObjectPreProcess(bucketName, objectName)
	if retCode != BOSCLI_OK {
		bcecliAbnormalExistCode(retCode)
	}

	// get storage class
	err := b.headObjectExecute(bucketName, objectName)
	if err != nil {
		bcecliAbnormalExistErr(err)
	}
}

// preprocessing
func (b *BosApi) headObjectPreProcess(bosPath string, objectPath string) (string,
		string, BosCliErrorCode) {

	retCode, err := checkBosPath(bosPath)
	if err != nil {
		return "", "", retCode
	}

	bucketName, objectKey := splitBosBucketKey(bosPath)
	if bucketName == "" {
		return "", "", BOSCLI_BUCKETNAME_IS_EMPTY
	} else if objectKey != "" {
		return "", "", BOSCLI_BUCKETNAME_CONTAIN_OBJECTNAME
	}

    objectName := strings.TrimSpace(objectPath)
	if objectName == "" {
		return "", "", BOSCLI_OBJECTKEY_IS_EMPTY
	}

	return bucketName, objectName, BOSCLI_OK
}

func (b *BosApi) headObjectExecute(bucketName, objectName string) error {
	// get object info
	ret, err := b.bosClient.GetObjectMeta(bucketName, objectName)
	if err != nil {
		return err
	}

	// print object information
	fmt.Printf(ret.GoString())
	return nil
}

