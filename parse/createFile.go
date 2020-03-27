/*
 * Copyright 2020 myaudit Author. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parse

import (
	"os"
	"strings"
	"strconv"
	"github.com/dbaxg/sqlpump/log"
)

/*

创建Mybatis Project所需的基础文件

 */

// 创建java bean
func createBean(labelInfo []string, pathBean string, columnList []string, packageName string) error {
	fileBean, err := os.Create(pathBean)
	defer fileBean.Close()
	if err != nil {
		log.LogIfError(err, "")
	}

	fileBean.WriteString(packageName + "\n")
	fileBean.WriteString("\n")
	for _, n := range columnList {
		if strings.Contains(n, "collection") {
			fileBean.WriteString("import java.util.List;\n")
			fileBean.WriteString("\n")
			break
		}
	}
	fileBean.WriteString("public class " + labelInfo[0] + "{\n")
	for _, n := range columnList {
		if strings.Contains(n, "collection") {
			fileBean.WriteString("private List " + n + ";\n")
		} else {
			fileBean.WriteString("private String " + n + ";\n")
		}
	}
	fileBean.WriteString("\n")
	for _, n := range columnList {
		if strings.Contains(n, "collection") {
			fileBean.WriteString("public List get" + strings.ToUpper(string(n[0])) + n[1:] + "() {\n")
			fileBean.WriteString("return " + n + ";\n")
			fileBean.WriteString("}\n")
			fileBean.WriteString("public void set" + strings.ToUpper(string(n[0])) + n[1:] + "(List " + n + "_P" + ")" + "{\n")
			fileBean.WriteString("this." + n + " = " + n + "_P;\n")
			fileBean.WriteString("}\n")
			fileBean.WriteString("\n")
		} else {
			fileBean.WriteString("public String get" + strings.ToUpper(string(n[0])) + n[1:] + "() {\n")
			fileBean.WriteString("return " + n + ";\n")
			fileBean.WriteString("}\n")
			fileBean.WriteString("public void set" + strings.ToUpper(string(n[0])) + n[1:] + "(String " + n + "_P" + "){\n")
			fileBean.WriteString("this." + n + " = " + n + "_P;\n")
			fileBean.WriteString("}\n")
			fileBean.WriteString("\n")
		}
	}
	_, err = fileBean.WriteString("}\n")
	return err
}

// 创建测试程序
func createTestProcess(labelId string, combined_list [][]string, pathMybatis string, packageName string) error {
	path := pathMybatis + "/" + "T_" + labelId + ".java"
	fileTest, err := os.Create(path)
	defer fileTest.Close()
	if err != nil {
		log.LogIfError(err, "")
	}

	fileTest.WriteString(packageName + "\n")
	fileTest.WriteString("\n")
	for _, t := range combined_list[0] {
		if strings.Contains(t, "collection") {
			fileTest.WriteString("import java.util.ArrayList;\n")
			fileTest.WriteString("import java.util.List;\n")
			break
		}
	}
	fileTest.WriteString("import org.apache.ibatis.session.SqlSession;\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("public class " + "T_" + labelId + "{\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("public static void main(String[] args) {\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("SqlSession session = DBTools.getSession();\n")
	fileTest.WriteString("newMapper mapper = session.getMapper(newMapper.class);\n")
	fileTest.WriteString("\n")
	for _, t := range combined_list[0] {
		if strings.Contains(t, "collection") {
			fileTest.WriteString("List<String> collection = new ArrayList<String>();\n")
			fileTest.WriteString("collection.add(\"SQLAudit\");\n")
			fileTest.WriteString("\n")
			break
		}
	}
	x := 0
	for _, y := range combined_list {
		fileTest.WriteString(labelId + " " + labelId + "_P" + strconv.Itoa(x) + " = " + "new " + labelId + "();\n")
		for _, z := range y {
			if strings.Contains(z, "collection") {
				fileTest.WriteString(labelId + "_P" + strconv.Itoa(x) + "." + "set" + strings.ToUpper(string(z[0])) + z[1:] + "(" + "collection" + ");\n")
			} else {
				fileTest.WriteString(labelId + "_P" + strconv.Itoa(x) + "." + "set" + strings.ToUpper(string(z[0])) + z[1:] + "(" + "\"SQLAudit\"" + ");\n")
			}
		}
		fileTest.WriteString(labelId + "(" + labelId + "_P" + strconv.Itoa(x) + ", session, mapper);\n")
		fileTest.WriteString("\n")
		x++
	}
	fileTest.WriteString("session.close();\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("}\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("private static void " + labelId + "(" + labelId + " " + labelId + "_P, SqlSession session, newMapper mapper) {\n")
	fileTest.WriteString("\n")
	fileTest.WriteString("try {\n")
	fileTest.WriteString("mapper." + labelId + "(" + labelId + "_P" + ");\n")
	fileTest.WriteString("session.commit();\n")
	fileTest.WriteString("} " + "catch (Exception e) {\n")
	fileTest.WriteString("session.rollback();\n")
	fileTest.WriteString("}\n")
	fileTest.WriteString("}\n\n")
	_, err = fileTest.WriteString("}\n")
	return err
}

// 创建配置文件
func createConf(pathMybatis string, xmlName string, testDb string, username string, password string) error {
	fileConf, err := os.Create(pathMybatis + "/conf.xml")
	defer fileConf.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	fileConf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fileConf.WriteString("<!DOCTYPE configuration PUBLIC \"-//mybatis.org//DTD Config 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-config.dtd\">\n")
	fileConf.WriteString("\n")
	fileConf.WriteString("<configuration>\n")
	fileConf.WriteString("   <environments default=\"development\">\n")
	fileConf.WriteString("           <environment id=\"development\">\n")
	fileConf.WriteString("                   <transactionManager type=\"JDBC\" />\n")
	fileConf.WriteString("                   <dataSource type=\"POOLED\">\n")
	fileConf.WriteString("                           <property name=\"driver\" value=\"com.mysql.jdbc.Driver\" />\n")
	fileConf.WriteString("                           <property name=\"url\" value=\"jdbc:mysql://" + testDb + "?characterEncoding=UTF-8\"/>\n")
	fileConf.WriteString("                           <property name=\"username\" value=\"" + username + "\"/>\n")
	fileConf.WriteString("                           <property name=\"password\" value=\"" + password + "\"/>\n")
	fileConf.WriteString("                   </dataSource>\n")
	fileConf.WriteString("           </environment>\n")
	fileConf.WriteString("   </environments>\n")
	fileConf.WriteString("\n")
	fileConf.WriteString("   <mappers>\n")
	fileConf.WriteString("           <mapper class=\"" + "SQLAudit." + xmlName + ".newMapper\" />\n")
	fileConf.WriteString("   </mappers>\n")
	fileConf.WriteString("\n")
	_, err = fileConf.WriteString("</configuration>")
	return err
}

// 创建log4 properties
func createLog4jProperties(pathMybatis string, xmlName string) error {
	file_log4j_properties, err := os.Create(pathMybatis + "/log4j.properties")
	defer file_log4j_properties.Close()
	if err != nil {
		log.LogIfError(err, "%s")
	}
	file_log4j_properties.WriteString("### set log levels ###\n")
	file_log4j_properties.WriteString("log4j.rootLogger = INFO, stdConsole, stdFile\n")
	file_log4j_properties.WriteString("log4j.logger.com.paic.dbaudit = DEBUG\n")
	file_log4j_properties.WriteString("\n")
	file_log4j_properties.WriteString("### set trace###\n")
	file_log4j_properties.WriteString("log4j.logger.SQLAudit." + xmlName + " = TRACE\n")
	file_log4j_properties.WriteString("\n")
	file_log4j_properties.WriteString("### output to the console ###\n")
	file_log4j_properties.WriteString("log4j.appender.stdConsole = org.apache.log4j.ConsoleAppender\n")
	file_log4j_properties.WriteString("log4j.appender.stdConsole.layout = org.apache.log4j.PatternLayout\n")
	file_log4j_properties.WriteString("log4j.appender.stdConsole.layout.ConversionPattern = %-d{yyyy-MM-dd HH:mm:ss} [%c.%M:%L]-[%p] %m%n\n")
	file_log4j_properties.WriteString("\n")
	file_log4j_properties.WriteString("### output to the log file ###\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile = org.apache.log4j.DailyRollingFileAppender\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile.layout = org.apache.log4j.PatternLayout\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile.File = " + pathMybatis + "/" + xmlName + ".log\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile.DatePattern='.'yyyy-MM-dd\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile.Append = true\n")
	file_log4j_properties.WriteString("log4j.appender.stdFile.Threshold = debug\n")
	_, err = file_log4j_properties.WriteString("log4j.appender.stdFile.layout.ConversionPattern = %-d{yyyy-MM-dd HH:mm:ss} [%c.%M:%L]-[%p] %m%n\n")
	return err
}

// 创建DBTools
func createDBTools(pathDir string, packageName string) error {
	fileDBTools, err := os.Create(pathDir + "/DBTools.java")
	defer fileDBTools.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	fileDBTools.WriteString(packageName + "\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("import java.io.Reader;\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("import org.apache.ibatis.io.Resources;\n")
	fileDBTools.WriteString("import org.apache.ibatis.session.SqlSession;\n")
	fileDBTools.WriteString("import org.apache.ibatis.session.SqlSessionFactory;\n")
	fileDBTools.WriteString("import org.apache.ibatis.session.SqlSessionFactoryBuilder;\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("public class DBTools {\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("    public static SqlSessionFactory sessionFactory;\n")
	fileDBTools.WriteString("    static{\n")
	fileDBTools.WriteString("        try {\n")
	fileDBTools.WriteString("            Reader reader = Resources.getResourceAsReader(\"conf.xml\");\n")
	fileDBTools.WriteString("            sessionFactory = new SqlSessionFactoryBuilder().build(reader);\n")
	fileDBTools.WriteString("        } catch (Exception e) {\n")
	fileDBTools.WriteString("            e.printStackTrace();\n")
	fileDBTools.WriteString("        }\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("    }\n")
	fileDBTools.WriteString("\n")
	fileDBTools.WriteString("    public static SqlSession getSession(){\n")
	fileDBTools.WriteString("        return sessionFactory.openSession();\n")
	fileDBTools.WriteString("    }\n")
	fileDBTools.WriteString("\n")
	_, err = fileDBTools.WriteString("}")
	return err
}

// 创建接口文件
func createInterface(idList []string, pathDir string, packageName string) error {
	fileInterface, err := os.Create(pathDir + "/newMapper.java")
	defer fileInterface.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	fileInterface.WriteString(packageName + "\n")
	fileInterface.WriteString("\n")
	fileInterface.WriteString("public interface newMapper {\n")
	for _, m := range idList {
		fileInterface.WriteString("void " + m + "(" + m + " " + m + "_P" + ");\n")
	}
	_, err = fileInterface.WriteString("}\n")
	return err
}

// 创建编译和运行Mybatis Project的shell脚本
func createSh(pathMybatis string, pathLib string, idList []string, xmlName string) error {
	fileSh, err := os.Create(pathMybatis + "/" + xmlName + ".sh")
	defer fileSh.Close()
	if err != nil {
		log.LogIfError(err, "")
	}
	fileSh.WriteString("\n")
	fileSh.WriteString("echo \"Compiling java files...\"\n")
	fileSh.WriteString("javac -d " + pathMybatis + " -cp " + pathLib + "/mybatis-3.5.4.jar " + pathMybatis + "/DBTools.java\n")
	for _, a := range idList {
		fileSh.WriteString("javac -d " + pathMybatis + " " + pathMybatis + "/" + a + ".java\n")
	}
	fileSh.WriteString("javac -d " + pathMybatis + " -cp " + pathMybatis + " " + pathMybatis + "/newMapper.java\n")
	for _, b := range idList {
		fileSh.WriteString("javac -d " + pathMybatis + " -cp " + pathLib + "/mybatis-3.5.4.jar:" + pathMybatis + " " + pathMybatis + "/" + "T_" + b + ".java\n")
	}
	fileSh.WriteString("\n")
	fileSh.WriteString("echo \"Copying newMapper.xml to directory where .class files have been...\"\n")
	fileSh.WriteString("cp " + pathMybatis + "/newMapper.xml " + pathMybatis + "/SQLAudit/" + xmlName + "\n")
	fileSh.WriteString("\n")
	fileSh.WriteString("echo \"executing .class files and writing sql to " + xmlName + ".log...\"\n")
	fileSh.WriteString("cd " + pathMybatis + "\n")
	for _, c := range idList {
		_, err = fileSh.WriteString("java -cp " + pathLib + "/*:. SQLAudit." + xmlName + "." + "T_" + c + "\n")
	}
	return err
}
