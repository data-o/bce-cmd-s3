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

// .

package boscli

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"bcecmd/boscmd"
)

// DeleteObjectResult defines the result structure for deleting a single object.
type DeleteObjectResult struct {
	Key     string `json:"key"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DeleteMultipleObjectsResult defines the result structure for deleting multiple objects.
type DeleteMultipleObjectsResult struct {
	Errors []DeleteObjectResult `json:"errors"`
}

type CompletedPart struct {
	ETag       string
	PartNumber int64
}

func NewCompletedMultipartUpload(size int64) *s3.CompletedMultipartUpload {
	return &s3.CompletedMultipartUpload{
		Parts: make([]*s3.CompletedPart, size),
	}
}

func AddNewCompletedPart(multi *s3.CompletedMultipartUpload, part *CompletedPart, id int) {
	multi.Parts[id] = &s3.CompletedPart{
		ETag:       aws.String(part.ETag),
		PartNumber: aws.Int64(part.PartNumber),
	}
}

func GetErrorCode(err error) (boscmd.BosErrorCode, bool) {
	if serverErr, ok := err.(awserr.Error); ok {
		return boscmd.BosErrorCode(serverErr.Code()), true
	}
	return BOSCLI_EMPTY_CODE, false
}

func GetErrorCodeAndMessage(err error) (boscmd.BosErrorCode, string, bool) {
	if serverErr, ok := err.(awserr.Error); ok {
		return boscmd.BosErrorCode(serverErr.Code()), serverErr.Message(), true
	}
	return BOSCLI_EMPTY_CODE, "", false
}

// get error message from error
func getErrorMsg(err error) string {
	if err == nil {
		return ""
	}
	if serverErr, ok := err.(awserr.Error); ok {
		return serverErr.Message()
	}
	return err.Error()
}

// Interface for wrap go sdk
type bosClientInterface interface {
	HeadBucket(bucket string) error
	ListBuckets() (*s3.ListBucketsOutput, error)
	ListObjects(string, string, string, string, int) (*s3.ListObjectsOutput, error)
	PutBucket(string) (string, error)
	DeleteBucket(string) error
	GetBucketLocation(string) (string, error)
	BasicGeneratePresignedUrl(string, string, int) string
	DeleteMultipleObjectsFromKeyList(string, []string) (*DeleteMultipleObjectsResult, error)
	DeleteObject(string, string) error
	GetObjectMeta(string, string) (*s3.HeadObjectOutput, error)
	CopyObject(bucket, object, srcBucket, srcObject, storageClass string,
	) (*s3.CopyObjectOutput, error)
	BasicGetObjectToFile(string, string, string) error
	PutObjectFromFile(string, string, string, string) (string, error)
	PutBucketLifecycleFromString(string, string) error
	GetBucketLifecycle(bucket string) (*s3.GetBucketLifecycleOutput, error)
	DeleteBucketLifecycle(string) error
	PutBucketStorageclass(string, string) error
	GetBucketStorageclass(string) (string, error)
	PutBucketAclFromCanned(string, string) error
	PutBucketAclFromString(string, string) error
	GetBucketAcl(string) (*s3.GetBucketAclOutput, error)
	UploadPartCopy(string, string, string, string, string, string,
		int64) (*s3.CopyPartResult, error)
	UploadPartFromBytes(bucket, object, uploadId string, partNumber int, content []byte,
		input *s3.UploadPartInput) (string, error)
	InitiateMultipartUpload(string, string, string, string) (string, error)
	AbortMultipartUpload(bucket, object, uploadId string) error
	CompleteMultipartUploadFromStruct(string, string, string,
		*s3.CompletedMultipartUpload) (*s3.CompleteMultipartUploadOutput, error)
	GetObject(string, string, map[string]string, ...int64) (*s3.GetObjectOutput, error)
}
