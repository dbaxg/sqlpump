# 查看帮助

```bash
[go@myaudit ~]$ myaudit -h
version: myaudit-2.0
Usage: myaudit [-h] [-f filename] [-s connStr] [-u username] [-p password] [-c fileConf]
Example: myaudit -f mapperTest.xml -s 127.0.0.1:3306/sakila -u xxx -p xxx -c /usr/etc/myaudit.toml
Options:
   -h show the usage of myaudit ~
   -f file to parse ~
   -s $IP:$PORT/$DB, like 127.0.0.1:3306/sakila ~
   -u database username ~
   -p database password ~
   -c configuration file, default `/usr/etc/myaudit.toml`~
Tips: If you don't declare these parameters above, myaudit will use the parameters in the configuration file.
```

# 执行解析

```bash
[go@myaudit ~]$ myaudit -f /usr/local/mapper/[mapperTest.xml](https://github.com/dbaxg/myaudit/tree/master/doc/mapperTest.xml)
{
"resultCode": 0,
"sqlPath": "/usr/local/myaudit/sql/mapperTest-1585267231364",
"errorInfo": "",
"panicInfo": "",
"stackInfo": ""
}
```

## json响应信息参数解释：
resultCode有0,1,2三个值，可根据myaudit返回的json串中的resultCode判断本次解析是否成功:
1) resultCode为0时，表示myaudit解析成功，用户可根据json中的sqlPath获取解析出来的SQL
2) resultCode为1时，表示myaudit执行遇到已知报错（如用户、密码错误/文件不存在等），用户可根据errorInfo和stackInfo来定位错误
3) resultCode为2时，表示myaudit执行时遇到未知bug发生panic（希望永远不要出现。。。），需根据errInfo和stackInfo来定位bug

# 查看SQL

```bash
[go@myaudit ~]$ cd /usr/local/myaudit/sql/mapperTest-1585267231364
[go@myaudit mapperTest-1585267231364]$ ll
总用量 36
-rw-r--r--. 1 oracle oinstall 424 3月  27 08:01 dynamicChooseTest.sql
-rw-r--r--. 1 oracle oinstall 308 3月  27 08:01 dynamicDeleteTest.sql
-rw-r--r--. 1 oracle oinstall  84 3月  27 08:01 dynamicForeach1Test.sql
-rw-r--r--. 1 oracle oinstall  84 3月  27 08:01 dynamicForeachTest.sql
-rw-r--r--. 1 oracle oinstall 952 3月  27 08:01 dynamicIfTest.sql
-rw-r--r--. 1 oracle oinstall 766 3月  27 08:01 dynamicSetTest.sql
-rw-r--r--. 1 oracle oinstall 782 3月  27 08:01 dynamicTrimTest.sql
-rw-r--r--. 1 oracle oinstall 788 3月  27 08:01 dynamicWhereTest.sql
-rw-r--r--. 1 oracle oinstall 146 3月  27 08:01 selectByLike.sql
[go@myaudit mapperTest-1585267231364]$ more dynamicChooseTest.sql
--select * from t_blog where 1 = 1 and title = ?;
  select * from t_blog where 1 = 1 and title = 'a';
--select * from t_blog where 1 = 1 and content = ?;
  select * from t_blog where 1 = 1 and content = 'a';
--select * from t_blog where 1 = 1 and title = ?;
  select * from t_blog where 1 = 1 and title = 'a';
--select * from t_blog where 1 = 1 and owner = "owner1";
  select * from t_blog where 1 = 1 and owner = "owner1";
```

## 补充说明：
1. 当动态SQL中的变量个数小于等于8个时，myaudit会对变量进行组合，然后传参，以提取出SQL所有可能的形态。
   当动态SQL中的变量个数大于8个时，myaudit只会对所有变量进行传参，提取出动态SQL最全面的那种形态。
   这是因为动态变量个数太多时，组合情况会很多（2的n次方，n为动态变量个数），出于性能考虑做了限制。
   用户可以通过修改源码来解除限制：[parse.go:215]

2. sqlPath下的sql会以label id命名，文件中'--'开头的SQL为myaudit通过执行自定义MyBatis Project解析出来的SQL，
   下面的SQL为myaudit根据字段类型进行变量替换后，生成的可执行SQL。