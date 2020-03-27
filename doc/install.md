## 源码安装

### 依赖软件
一般依赖

* Go 1.10+
* Java 1.8+

### 生成二进制文件

```bash
cd ${GOPATH} && git clone https://github.com/dbaxg/sqlpump.git
cd sqlpump/cmd && go build sqlpump.go
```

## 安装验证

```bash
./sqlpump -h
```