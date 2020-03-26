## 源码安装

### 依赖软件
一般依赖

* Go 1.10+
* Java 1.8+

### 生成二进制文件

```bash
go get -d github.com/dbaxg/myaudit
cd ${GOPATH}/src/github.com/dbaxg/cmd && go build myaudit.go
```

## 安装验证

```bash
./myaudit -h
```