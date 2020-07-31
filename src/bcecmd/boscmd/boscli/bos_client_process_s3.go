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

// Codes in this file is used to process bosClient.

package boscli

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"bceconf"
)

const (
	defaultMaxIdleConnsPerHost   = 1000
	defaultMaxIdleConns          = 1000
	defaultIdleConnTimeout       = 0
	defaultResponseHeaderTimeout = 60 * time.Second
	defaultDialTimeout           = 30 * time.Second
)

func getUserAgent() string {
	var userAgent string
	userAgent += BCE_CLI_AGENT
	userAgent += "/" + bceconf.BCE_VERSION
	userAgent += "/" + runtime.Version()
	userAgent += "/" + runtime.GOOS
	userAgent += "/" + runtime.GOARCH
	return userAgent
}

// build a new http client
func newHttpClient() *http.Client {
	httpClient := &http.Client{}
	transport := &http.Transport{
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		ResponseHeaderTimeout: defaultResponseHeaderTimeout,
		IdleConnTimeout:       defaultIdleConnTimeout,
		Dial: func(network, address string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, address, defaultDialTimeout)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	httpClient.Transport = transport
	return httpClient
}

func setEndpointProtocol(endpoint string, useHttps bool) (string, error) {
	if strings.HasPrefix(endpoint, HTTP_PROTOCOL) {
		endpoint = endpoint[len(HTTP_PROTOCOL):]
	} else if strings.HasPrefix(endpoint, HTTPS_PROTOCOL) {
		endpoint = endpoint[len(HTTPS_PROTOCOL):]
	}
	if endpoint == "" {
		return "", fmt.Errorf("Endpoint is empty!")
	}
	if useHttps {
		return HTTPS_PROTOCOL + endpoint, nil
	}
	return HTTP_PROTOCOL + endpoint, nil
}

func newBosClient(ak, sk, stsToken, endpoint string, useHttps bool) (*s3ClientWrapper, error) {

	// set http or https protocol
	endpoint, err := setEndpointProtocol(endpoint, useHttps)
	if err != nil {
		return nil, fmt.Errorf("Endpoint is invalid!")
	}

	cres := credentials.NewStaticCredentials(ak, sk, stsToken)
	region, _ := bceconf.ServerConfigProvider.GetRegion()
	cfg := &aws.Config{
		Credentials: cres,
		Endpoint:    &endpoint,
		Region:      &region,
		HTTPClient:  newHttpClient(),
	}
	cfg = cfg.WithS3ForcePathStyle(true).WithS3Disable100Continue(true)
	if bceconf.DebugLevel {
		cfg.WithLogLevel(aws.LogDebugWithRequestRetries)
		//cfg.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}

	s3Client := &s3ClientWrapper{
		s3Client: s3.New(sess),
	}

	return s3Client, nil
}

// Init bos client.
func buildBosClient(ak, sk, endpoint string,
	credentialProvider bceconf.CredentialProviderInterface,
	serverConfigProvider bceconf.ServerConfigProviderInterface) (*s3ClientWrapper, error) {
	var (
		ok bool
	)

	if ak == "" || sk == "" {
		if ak, ok = credentialProvider.GetAccessKey(); !ok {
			return nil, fmt.Errorf("There is no access key found!")
		}
		if sk, ok = credentialProvider.GetSecretKey(); !ok {
			return nil, fmt.Errorf("There is no access secret key found!")
		}
	}

	stsToken, _ := credentialProvider.GetSecurityToken()

	if endpoint == "" {
		if endpoint, ok = serverConfigProvider.GetDomain(); !ok {
			return nil, fmt.Errorf("There is no endpoint found!")
		}
	}

	if useHttps, ok := serverConfigProvider.GetUseHttpsProtocol(); !ok {
		return nil, fmt.Errorf("There is no https protocol info found!")
	} else {
		return newBosClient(ak, sk, stsToken, endpoint, useHttps)
	}
}

// init BosCLient
func bosClientInit(ak, sk, endpoint string) (bosClientInterface, error) {
	return buildBosClient(
		ak,
		sk,
		endpoint,
		bceconf.CredentialProvider,
		bceconf.ServerConfigProvider,
	)
}
