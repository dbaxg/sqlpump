## MYAUDIT

MYAUDIT是一个对MyBatis xml文件进行SQL提取和变量回填的自动化工具，可用于SQL前置审核或代码库审核。

## 功能特点

* 与业务深度解耦，只需上传xml文件并配置开发环境的数据库信息
* 支持提取动态SQL所有组合情况下的SQL指纹
* 支持对SQL指纹按变量字段类型进行填值，以获得完整SQL，便于后续执行计划的获取和SQL的审核
* 支持 UPDATE, DELETE, SELECT等类型SQL的提取
* 目前只支持 MySQL 语法族的解析和变量回填
* 支持json格式的响应信息，方便外部程序的调用

## 快速入门

* [安装使用](https://github.com/dbaxg/myaudit/tree/master/doc/install.md)
* [体系架构](https://github.com/dbaxg/myaudit/tree/master/doc/structure.md)
* [配置文件](https://github.com/dbaxg/myaudit/tree/master/doc/config.md)
## License

[Apache License 2.0](https://github.com/dbaxg/myaudit/tree/master/LICENSE).