WORKROOT=$(pwd)
export GOPATH=${WORKROOT}

# try to download bce-sdk-go from github
go env -w GO111MODULE=on
if [ $? -ne 0 ]
then
    echo "fail to set env GO111MODULE"
    exit 1
fi

go env -w GONOPROXY=\*\*.baidu.com\*\*
go env -w GOPROXY=https://goproxy.baidu.com
go env -w GONOSUMDB=\*


chmod +w pkg -R 2>&1 > /dev/null
rm -rf pkg/mod/github.com/*
rm -rf ./src/github.com/*

echo "start to get bce-sdk-go"
go get -d github.com/baidubce/bce-sdk-go@1a69080
if [ $? -ne 0 ]
then
    echo "fail to get bce-sdk-go"
    exit 1
fi

echo "start to get aws-sdk-go"
go get -d github.com/aws/aws-sdk-go@v1.28.13
if [ $? -ne 0 ]
then
    echo "fail to get aws-sdk-go"
    exit 1
fi

echo "start to get go-jmespath"
go get -d github.com/jmespath/go-jmespath@c2b33e8
if [ $? -ne 0 ]
then
    echo "fail to get go-jmespath"
    exit 1
fi

chmod +w ./pkg -R
mkdir -p ./src/github.com/baidubce
mkdir -p ./src/github.com/aws
mkdir -p ./src/github.com/jmespath
mv pkg/mod/github.com/baidubce/bce-sdk-go* ./src/github.com/baidubce/bce-sdk-go
mv pkg/mod/github.com/aws/aws-sdk-go* ./src/github.com/aws/aws-sdk-go
mv pkg/mod/github.com/jmespath/go-jmespath* ./src/github.com/jmespath/go-jmespath

go env -w GO111MODULE=off

# start to build bcecmd
cd $WORKROOT/src/main
go build
if [ $? -ne 0 ];
then
	echo "fail to build bcecmd.go"
	exit 1
fi

cd ../../
if [ -d "./output" ]
then
	rm -rf output
fi

mkdir output
mv src/main/main "output/bcecmds3"
chmod +x output/bcecmds3

echo "OK for build bcecmds3"
