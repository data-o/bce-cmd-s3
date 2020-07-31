# BOSCMDS3

我们基于百度云开源的 [BCE CMD](https://github.com/baidubce/bce-cmd) 添加了对 S3 的支持。

## 快速开始

### 编译

编译前提：

1. 安装 golang (至少为 1.11);
2. 配置环境变量 GOROOT；

执行下面的代码，快速编译：

```
sh build.sh
```

编译产出在目录 ./output 中.

### 配置和使用 bcecmd

请参考 [BCECMD官方文档](https://cloud.baidu.com/doc/BOS/s/Sjwvyqetg)

## 测试

```
go test -v $1 -coverprofile=c.out #请将 $1 替换你需要测试的文件或目录
go tool cover -html=c.out -o cover.html
```

## 如何贡献

迎您修改和完善此项目，请直接提交PR 或 issues。

* 提交代码时请保证良好的代码风格。
* 提交 issues 时， 请翻看历史 issues， 尽量不要提交重复的issues。

## 讨论

欢迎提 issues。
