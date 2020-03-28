## SQLPUMP

SQLPUMP是一个对MyBatis xml文件进行SQL抽取和变量回填的自动化工具。

## 功能特点

* 与业务代码深度解耦，只需上传xml文件即可完成SQL抽取
* 支持抽取动态SQL所有组合情况下的SQL指纹
* 支持对SQL指纹按变量字段类型进行填值，以获得可执行的完整SQL，便于后续的SQL审核
* 支持 UPDATE, DELETE, SELECT等类型SQL的提取
* 目前只支持MySQL语法族的解析和变量回填
* 支持json格式的响应信息，方便外部程序的调用

## 快速入门
* [安装使用](https://github.com/dbaxg/sqlpump/tree/master/doc/install.md)
* [设计思路](https://github.com/dbaxg/sqlpump/tree/master/doc/structure.md)
* [配置文件](https://github.com/dbaxg/sqlpump/tree/master/doc/config.md)
* [操作指南](https://github.com/dbaxg/sqlpump/tree/master/doc/handbook.md)

## 使用场景

#### 如果你是一名DBA或者运维
* 你可以将sqlpump引入你们的DB运维平台，通过sqlpump来全量提取xml文件中的动态SQL，在代码上线前给出优化建议，降低风险SQL的概率。
* 你可以通过sqlpump来扫描代码库（包括历史代码），来提取SQL并生成代码质量报告，将未来或者历史遗留的风险SQL揪出来。

#### 如果你是一名Java开发人员
* 你可以在代码开发阶段使用sqlpump来全量提取xml文件中的动态SQL，并用soar等工具来生成优化报告，帮助你发现代码中的缺陷，写出更高效的SQL。

#### 如果你是一名测试人员
* 你可以通过sqlpump来全量提取xml文件中的动态SQL，对MyBatis xml中的SQL进行360°覆盖测试。

## 案例展示

以test.xml为例：
```bash
[go@sqlpump ~]$ more /usr/local/mapper/test.xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<mapper namespace="any.namespace.is.ok">

<select id="dynamicIfTrimTest" parameterType="Blog" resultType="Blog"><!--
                            sqlpump与业务代码深度解耦，只需上传无差别的xml文件即可。
                            连接数据库的目的是为了查询information_schema.columns表，根据字段类型给动态变量赋值，
                            后续的版本中，会考虑把变量赋值功能独立，通过参数来控制是否进行变量赋值
                            -->
    select * from t_blog
    <trim prefix="where" prefixOverrides="and |or">
                id in
                <foreach collection="list" index="index" item="item" open="(" separator="," close=")">
        #{item}
                </foreach>
        <if test="title != null">
            and title = #{title}
        </if>
        <if test="content != null">
            and content = #{content}
        </if>
        <if test="owner != null">
            or owner = #{owner}
        </if>
    </trim>
</select>

</mapper>
```

小试牛刀：
```bash
[go@sqlpump ~]$ sqlpump -f /usr/local/mapper/test.xml
{
"resultCode": 0,
"sqlPath": "/usr/local/sqlpump/sql/test-1585306149670",
"errorInfo": "",
"panicInfo": "",
"stackInfo": ""
}
```

**test.xml中，id为dynamicIfTrimTest的动态select语句的所有可能出现的SQL均被抽取出，并根据字段类型进行了赋值：**
```bash
[go@sqlpump ~]$ cd /usr/local/sqlpump/sql/test-1585306149670
[go@sqlpump ~]$ ll
总用量 4
-rw-r--r--. 1 oracle oinstall 1040 3月  27 18:49 dynamicIfTrimTest.sql   --生成的.sql文件以标签id命名
[oracle@oracle-test dynamicIfTrimTest-1585306149670]$ more dynamicIfTrimTest.sql
--select * from t_blog where id in ( ? ) and title = ?;
  select * from t_blog where id in ( 1 ) and title = 'a';
--select * from t_blog where id in ( ? ) and content = ?;
  select * from t_blog where id in ( 1 ) and content = 'a';
--select * from t_blog where id in ( ? ) or owner = ?;
  select * from t_blog where id in ( 1 ) or owner = 'a';
--select * from t_blog where id in ( ? ) and title = ? and content = ?;
  select * from t_blog where id in ( 1 ) and title = 'a' and content = 'a';
--select * from t_blog where id in ( ? ) and title = ? or owner = ?;
  select * from t_blog where id in ( 1 ) and title = 'a' or owner = 'a';
--select * from t_blog where id in ( ? ) and content = ? or owner = ?;
  select * from t_blog where id in ( 1 ) and content = 'a' or owner = 'a';
--select * from t_blog where id in ( ? ) and title = ? and content = ? or owner = ?;
  select * from t_blog where id in ( 1 ) and title = 'a' and content = 'a' or owner = 'a';
--select * from t_blog where id in ( ? );
  select * from t_blog where id in ( 1 );
```

## License

[Apache License 2.0](https://github.com/dbaxg/sqlpump/tree/master/LICENSE)
