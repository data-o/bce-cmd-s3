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

// Interface for bos cli handler
type handlerInterface interface {
	multiDeleteDir(bosClientInterface, string, string) (int, error)
	multiDeleteObjectsWithRetry(bosClientInterface, []string, string) ([]DeleteObjectResult,
		error)
	utilDeleteObject(bosClientInterface, string, string) error
	utilCopyObject(bosClientInterface, bosClientInterface, string, string, string, string, string,
		int64, int64, int64, bool) error
	utilDownloadObject(bosClientInterface, string, string, string, string, bool, int64, int64, int64,
		bool) error
	utilUploadFile(bosClientInterface, string, string, string, string, string, int64, int64,
		int64, bool) error
	utilDeleteLocalFile(string) error
	doesBucketExist(bosClientInterface, string) (bool, error)
	CopySuperFile(bosClientInterface, bosClientInterface, string, string, string, string,
		string, int64, int64, int64, bool, string) error
}

// File Information: be used by BOS object and local file.
type fileDetail struct {
	path         string // both, full path (path of bos object don't contail bucket name)
	name         string // without replcae os sep to bos sep
	key          string // both
	realPath     string // local file, real path of symbolic link
	storageClass string // bos object
	crc32        string
	size         int64 // both
	mtime        int64 // both, last Modified time
	gtime        int64 // both, the time of get info of this object
	isDir        bool
	err          error // both
}

// Directory Information: be used by BOS pre.
type dirDetail struct {
	path string
	key  string
}

// End Information: be used by BOS.
type listEndInfo struct {
	nextMarker  string
	isTruncated bool
}

type listFileResult struct {
	file    *fileDetail
	dir     *dirDetail
	endInfo *listEndInfo
	err     error //have error when list object?
	isDir   bool  //is file or dir?
	ended   bool  //have get all object?
}

type fileListIterator interface {
	next() (*listFileResult, error)
}

type executeResult struct {
	failed    int
	successed int
}
