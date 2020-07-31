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

// This file provide functions to init and midify bos client.

package boscli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"utils/util"
)

const (
	PARALLEL_DELETE_NUM       = 50
	EACH_ROUTHINE_MIN_OBJECTS = 10
)

type s3ClientWrapper struct {
	s3Client *s3.S3
}

// Wrapper head bucket
func (b *s3ClientWrapper) HeadBucket(bucket string) error {
	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	_, err := b.s3Client.HeadBucket(input)
	return err
}

// Wrapper ListBuckets - list all buckets
func (b *s3ClientWrapper) ListBuckets() (*s3.ListBucketsOutput, error) {
	input := &s3.ListBucketsInput{}
	return b.s3Client.ListBuckets(input)
}

// Wrapper PutBucket - create a new bucket
func (b *s3ClientWrapper) PutBucket(bucket string) (string, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(""),
		},
	}
	ret, err := b.s3Client.CreateBucket(input)
	if err != nil {
		return "", err
	} else if ret.Location != nil {
		return *ret.Location, err
	} else {
		return "unknown", err
	}
}

// Wrapper GetBucketLocation - get the location fo the given bucket
func (b *s3ClientWrapper) GetBucketLocation(bucket string) (string, error) {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}
	ret, err := b.s3Client.GetBucketLocation(input)
	if err != nil {
		return "", err
	} else {
		return *ret.LocationConstraint, nil
	}
}

// Wrapper ListObjects - list all objects of the given bucket
func (b *s3ClientWrapper) ListObjects(bucket, delimiter, marker, prefix string,
	maxkeys int) (*s3.ListObjectsOutput, error) {

	input := &s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String(delimiter),
		Marker:    aws.String(marker),
		MaxKeys:   aws.Int64(int64(maxkeys)),
		Prefix:    aws.String(prefix),
	}

	return b.s3Client.ListObjects(input)
}

// Wrapper DeleteBucket - delete a empty bucket
func (b *s3ClientWrapper) DeleteBucket(bucket string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	}
	_, err := b.s3Client.DeleteBucket(input)
	return err
}

// Wrapper BasicGeneratePresignedUrl  generate an authorization url with expire time
func (b *s3ClientWrapper) BasicGeneratePresignedUrl(bucket string, object string,
	expireInSeconds int) string {
	return ""
}

// routineNum shouldn't be 0
func getObjectNumOfEachRoutine(total, routineNum, minNum int) int {
	each := total / routineNum
	if each <= minNum {
		return minNum
	}
	return each + 1
}

// Wrapper DeleteMultipleObjectsFromKeyList - delete a list of objects with given key string array
func (b *s3ClientWrapper) DeleteMultipleObjectsFromKeyList(bucket string,
	keyList []string) (*DeleteMultipleObjectsResult, error) {

	var (
		unDelLists []DeleteObjectResult
		opSync     sync.WaitGroup
	)

	objectNum := len(keyList)
	if objectNum == 0 {
		return nil, io.EOF
	}

	// init channel
	syncOpPool := make(chan int, PARALLEL_DELETE_NUM)
	executeResultChan := make(chan []DeleteObjectResult, PARALLEL_DELETE_NUM)

	// this function is used to execute sync operation
	cpOpFunc := func(bucket string, keyList []string, start, end int,
		retChan chan []DeleteObjectResult, wg *sync.WaitGroup) {
		var (
			fails []DeleteObjectResult
		)

		defer func() {
			wg.Done()
			<-syncOpPool
		}()

		retry := 0
		for i := start; i < end && retry < 3; i++ {
			input := &s3.DeleteObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(keyList[i]),
			}
			_, err := b.s3Client.DeleteObject(input)
			if err == nil {
				continue
			}
			if awsErr, ok := err.(awserr.Error); ok {
				retry++
				fails = append(fails, DeleteObjectResult{
					Key:     keyList[i],
					Code:    awsErr.Code(),
					Message: awsErr.Message(),
				})
			} else {
				retry++
				fails = append(fails, DeleteObjectResult{
					Key:     keyList[i],
					Code:    "None",
					Message: err.Error(),
				})
			}
		}
		retChan <- fails
	}

	// upload from file
	each := getObjectNumOfEachRoutine(objectNum, PARALLEL_DELETE_NUM, EACH_ROUTHINE_MIN_OBJECTS)
	run_routine := 0
	for start := 0; start < objectNum; start += each {
		end := start + each
		if end > objectNum {
			end = objectNum
		}
		opSync.Add(1)
		go cpOpFunc(bucket, keyList, start, end, executeResultChan, &opSync)
		run_routine++
	}

	for i := 0; i < run_routine; i++ {
		ret := <-executeResultChan
		if len(ret) != 0 {
			unDelLists = append(unDelLists, ret...)
		}
	}

	// waiting for all sync operation finish
	opSync.Wait()
	if len(unDelLists) == 0 {
		return nil, io.EOF
	}
	return &DeleteMultipleObjectsResult{
		Errors: unDelLists,
	}, fmt.Errorf("failed delete some objects")
}

// Wrapper DeleteObject - delete the given object
func (b *s3ClientWrapper) DeleteObject(bucket, object string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	_, err := b.s3Client.DeleteObject(input)
	return err
}

// Wrapper GetObjectMeta
func (b *s3ClientWrapper) GetObjectMeta(bucket, object string) (*s3.HeadObjectOutput, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	return b.s3Client.HeadObject(input)
}

// Wrapper Copy Object
func (b *s3ClientWrapper) CopyObject(bucket, object, srcBucket, srcObject, storageClass string,
) (*s3.CopyObjectOutput, error) {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		CopySource: aws.String(srcBucket + util.BOS_PATH_SEPARATOR + srcObject),
		Key:        aws.String(object),
	}
	if storageClass != "" {
		input.StorageClass = aws.String(storageClass)
	}
	return b.s3Client.CopyObject(input)
}

// Wrapper of BasiGetObjectToFile
func (b *s3ClientWrapper) BasicGetObjectToFile(bucket, object, localPath string) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	res, err := b.s3Client.GetObject(input)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	file, fileErr := os.OpenFile(localPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	written, writeErr := io.CopyN(file, res.Body, *res.ContentLength)
	if writeErr != nil {
		return writeErr
	}
	if written != *res.ContentLength {
		return fmt.Errorf("written content size does not match the response content")
	}
	return nil
}

// Wrapper of PutObjectFromFile
func (b *s3ClientWrapper) PutObjectFromFile(bucket, object, fileName, storageClass string) (string,
	error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	input := &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	res, err := b.s3Client.PutObject(input)
	if err != nil {
		return "", err
	}
	return *res.ETag, nil
}

// Wrapper of PutBucketAclFromCanned
func (b *s3ClientWrapper) PutBucketAclFromCanned(bucket, cannedAcl string) error {
	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(cannedAcl),
	}
	_, err := b.s3Client.PutBucketAcl(input)
	return err
}

// Wrapper of PutBucketAcl
func (b *s3ClientWrapper) PutBucketAclFromString(bucket string, acl string) error {
	return fmt.Errorf("not support !")
}

// Wrapper of GetBucketAcl
func (b *s3ClientWrapper) GetBucketAcl(bucket string) (*s3.GetBucketAclOutput, error) {
	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}
	return b.s3Client.GetBucketAcl(input)
}

// Wrapper of PutBucketLifecycleFromString
func (b *s3ClientWrapper) PutBucketLifecycleFromString(bucket, lifecycle string) error {
	return fmt.Errorf("not support !")
}

// Wrapper of GetBucketLifecycle
func (b *s3ClientWrapper) GetBucketLifecycle(bucket string) (*s3.GetBucketLifecycleOutput,
	error) {
	return nil, fmt.Errorf("not support !")
}

// Wrapper of DeleteBucketLifecycle
func (b *s3ClientWrapper) DeleteBucketLifecycle(bucket string) error {
	return fmt.Errorf("not support !")
}

// Wrapper of PutBucketStorageclass
func (b *s3ClientWrapper) PutBucketStorageclass(bucket, storageClass string) error {
	return fmt.Errorf("not support !")
}

// Wrapper of GetBucketStorageclass
func (b *s3ClientWrapper) GetBucketStorageclass(bucket string) (string, error) {
	return "", fmt.Errorf("not support !")
}

// Wrapper of GetBucketStorageclass
func (b *s3ClientWrapper) UploadPartCopy(bucket, object, srcBucket, srcObject, uploadId,
	sourceRange string, partNumber int64) (*s3.CopyPartResult, error) {

	input := &s3.UploadPartCopyInput{
		Bucket:          aws.String(bucket),
		CopySource:      aws.String(srcBucket + util.BOS_PATH_SEPARATOR + srcObject),
		CopySourceRange: aws.String(sourceRange),
		Key:             aws.String(object),
		PartNumber:      aws.Int64(partNumber),
		UploadId:        aws.String(uploadId),
	}

	res, err := b.s3Client.UploadPartCopy(input)
	if err != nil {
		return nil, err
	}
	return res.CopyPartResult, nil
}

// Wrapper of GetBucketStorageclass
func (b *s3ClientWrapper) UploadPartFromBytes(bucket, object, uploadId string, partNumber int,
	content []byte, input *s3.UploadPartInput) (string, error) {

	if input == nil {
		input = &s3.UploadPartInput{}
	}
	input.SetBody(bytes.NewReader(content))
	input.SetBucket(bucket)
	input.SetKey(object)
	input.SetPartNumber(int64(partNumber))
	input.SetUploadId(uploadId)
	res, err := b.s3Client.UploadPart(input)
	if err != nil {
		return "", err
	}
	return *res.ETag, nil
}

// Wrapper of GetBucketStorageclass
func (b *s3ClientWrapper) InitiateMultipartUpload(bucket, object, contentType,
	storageClass string) (string, error) {

	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}
	if contentType != "" {
		input.SetContentType(contentType)
	}
	if storageClass != "" {
		input.SetStorageClass(storageClass)
	}
	res, err := b.s3Client.CreateMultipartUpload(input)
	if err != nil {
		return "", err
	}
	return *res.UploadId, err
}

func (b *s3ClientWrapper) AbortMultipartUpload(bucket, object, uploadId string) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadId: aws.String(uploadId),
	}
	_, err := b.s3Client.AbortMultipartUpload(input)
	return err
}

func (b *s3ClientWrapper) CompleteMultipartUploadFromStruct(bucket, object, uploadId string,
	parts *s3.CompletedMultipartUpload) (*s3.CompleteMultipartUploadOutput, error) {

	input := &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(object),
		MultipartUpload: parts,
		UploadId:        aws.String(uploadId),
	}
	return b.s3Client.CompleteMultipartUpload(input)
}

func (b *s3ClientWrapper) GetObject(bucket, object string, responseHeaders map[string]string,
	ranges ...int64) (*s3.GetObjectOutput, error) {

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}

	if len(ranges) != 0 {
		rangeStr := "bytes="
		if len(ranges) == 1 {
			rangeStr += fmt.Sprintf("%d", ranges[0]) + "-"
		} else {
			rangeStr += fmt.Sprintf("%d", ranges[0]) + "-" + fmt.Sprintf("%d", ranges[1])
		}
		input.SetRange(rangeStr)
	}
	return b.s3Client.GetObject(input)
}
