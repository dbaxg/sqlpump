## 配置文件说明

配置文件为[toml]格式。只需要配置FileName, UserName, Password, TestDSN和PathRoot共5个参数。


默认文件为`/etc/sqlpump.toml`。如需指定其他配置文件可以通过`-c`参数指定。

```text
# configuration template for sqlpump

[parm]
# MyBatis xml文件名
Filename = "mapperTest.xml"
# 数据库用户，只需最基本的usage权限，无需任何额外权限
Username = "xxx"
# 数据库用户密码
Password = "xxx"
# 非线上环境数据库连接串
TestDSN = "127.0.0.1:3306/sakila"

[path]
# 根路径用于存放sqlpump执行抽取的过程中生成的各类文件和项目依赖，需确保根路径存在且具备读写权限
PathRoot = "/usr/local/sqlpump"

```

## 命令行参数

所有配置文件中指定的参数均可通过命令行参数进行修改，且命令行参数优先级较配置文件优先级高。

```bash
sqlpump -h
version: sqlpump-1.0
Usage: sqlpump [-h] [-f filename] [-s connStr] [-u username] [-p password] [-c fileConf]
Example: sqlpump -f mapperTest.xml -s 127.0.0.1:3306/sakila -u xxx -p xxx -c /usr/etc/sqlpump.toml
Options:
   -h show the usage of sqlpump ~
   -f file to parse ~
   -s $IP:$PORT/$DB, like 127.0.0.1:3306/sakila ~
   -u database username ~
   -p database password ~
   -c configuration file, default `/etc/sqlpump.toml`~
Tips: If you don't declare these parameters above, sqlpump will use the parameters in the configuration file.
```
