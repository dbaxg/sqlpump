``` ![image](https://github.com/gaozhongzheng/myaudit/blob/master/doc/大黄.png)
### myaudit是一个审计SSM + MySQL开发架构中SQL映射文件（xml文件）的SQLAudit工具。
设计思路如下：
* 1.格式化xml文件
* 2.重写xml文件
* 3.构造Mybatis工程
* 4.调起Mybatis工程打印所有可能的SQL到日志
* 5.解析日志获取SQL，并根据动态字段类型将SQL中的'?'替换为具体值
* 6.调用SOAR获取审计报告
### 特点：
* 1.与业务代码深度解耦
* 2.解析并审计动态SQL所有可能的情况
* 3.根据字段类型替换绑定变量值，解决了SQL指纹无法在MySQL中获取执行计划的问题
### 安装/部署（Linux）
#### 方式一
* step1.克隆源代码至本地
* step2.编译命令文件：cd cmd;go build myaudit.go
* step3.执行[createDir.sh](https://github.com/gaozhongzheng/myaudit/blob/master/doc/createDir.sh)创建相应目录，目录含义见[目录说明](https://github.com/gaozhongzheng/myaudit/blob/master/doc/dirDescription.md)。
* step4.上传log4j-1.2.17.jar、mybatis-3.2.8.jar、mysql-connector-java-5.1.47.jar至/usr/local/myaudit/lib，可从[此处](https://github.com/gaozhongzheng/deployment/tree/master/lib)获取jar包
* step5.将step2中编译好的命令文件拷贝至/usr/local/myaudit/bin并将此目录添加至环境变量
##### 注：需先安装[SOAR](https://github.com/XiaoMi/soar)
#### 方式二(免编译快速部署)
* step1.克隆[deployment](https://github.com/gaozhongzheng/deployment)至本地
* step2.将/usr/local/myaudit/bin添加至环境变量
### 使用myaudit（如有疑问或建议，请邮件至2818962342@qq.com）
~~~
Usage of myaudit:
    -ftype       mapper|slowlog|normal, default mapper          --待审计文件种类，默认为mapper（暂时只支持Mybatis项目中的mapper文件）
    -fname       file waitting for audit ~                      --待审计文件名，不需加扩展名，如MapperSakila.xml传参时为MapperSakila
    -conn        $IP:$PORT/$DB, like 127.0.0.1:3306/sakila ~    --测试环境连接串，待审计文件对应的测试环境数据库
    -u           username    --MySQL用户名
    -p           password    --MySQL用户密码
    -report-type html|json, default html    --报告类型，默认为html
  ~~~
* step1.上传待审计的SQL映射文件至/usr/local/myaudit/file
* step2.执行审计（示例）：
~~~
myaudit -ftype mapper -fname [MapperSakila](https://github.com/gaozhongzheng/deployment/blob/master/file/MapperSakila.xml) -conn 127.0.0.1:3306/sakila -u root -p 123456 -report-type html
Dir '/usr/local/myaudit/tmp/MapperSakila' was created/rebuilt successfully!
/usr/local/myaudit/file/MapperSakila.xml read ok!
/usr/local/myaudit/tmp/MapperSakila/tmp.xml read ok!
/usr/local/myaudit/tmp/MapperSakila/oldMapper.xml read ok!
/usr/local/myaudit/tmp/MapperSakila/oldMapperFormated.xml read ok!
Compile and execute java files successfully!
/usr/local/myaudit/log/mapperLog/MapperSakila.log read ok!
Database connect success!
Dir '/usr/local/myaudit/sql/MapperSakila' was created/rebuilt successfully!
Dir '/usr/local/myaudit/audit/MapperSakila' was created/rebuilt successfully!
xml file was successfully audited, you can find the audit file 'getFilmInfoTest1_audit.html' in directory /usr/local/myaudit/audit/MapperSakila.
xml file was successfully audited, you can find the audit file 'getFilmInfoTest2_audit.html' in directory /usr/local/myaudit/audit/MapperSakila.
~~~
从[MapperSakila.xml](https://github.com/gaozhongzheng/deployment/blob/master/file/MapperSakila.xml)解析出的SQL文件[在此](https://github.com/gaozhongzheng/deployment/tree/master/sql/MapperSakila)。
#### 审计报告截图：
![audit file](https://github.com/gaozhongzheng/myaudit/blob/master/doc/MapperSakila_audit.png)

```
