package magic

import (
    "bufio"
    "database/sql"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strconv"
    "strings"

    "github.com/myaudit/common"
)

/**
    暂不支持xml文件中同一对引号不在同一行的情况，如：
    <select id="xxx"  resultType="xxx.
    xxx">
**/

// Makedir 创建目录
func Makedir(path string) error {

    isExist, err := func(path string) (bool, error) {
        _, err := os.Stat(path)
        if err == nil {
            return true, nil
        }
        if os.IsNotExist(err) {
            return false, nil
        }
        return false, err
    }(path)

    if isExist && err == nil {

        err = os.RemoveAll(path)
        if err != nil {
            return err
        }
        err = os.Mkdir(path, os.ModePerm)
        if err != nil {
            return err
        }

    }

    if !isExist && err == nil {
        err = os.Mkdir(path, os.ModePerm)
        if err != nil {
            return err
        }
    }
    return err
}

func check(e error) {
    if e != nil {
        common.LogIfError(e, "")
        panic(e)
    }
}

// Filter 获取Mapper文件中的select部分
func Filter(path_dir string, xml_input string, xml_output string, xml_name string) {
    xml_tmp := path_dir + "/tmp.xml"
    file_xml, err := os.Open(xml_input)
    check(err)
    defer file_xml.Close()

    file_xml_tmp_w, err := os.Create(xml_tmp)
    check(err)
    defer file_xml_tmp_w.Close()

    /**
        第一次换行的原因是防止下述情况：
        <select id...>
        select
        a            <!--
                      注释
                     -->
        from t
        </select>
        如果不换行，a所在行会被忽略，所以必须换行为：
        <select id...>
        select
        a
        <!--
                      注释
                     -->

        from t
        </select>
    */

    buf := bufio.NewReader(file_xml)
    for {
        line, err := buf.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println(xml_input + " read ok!")
                break
            } else {
                fmt.Println(xml_input+" read error!", err)
                common.LogIfError(err, "")
                return
            }
        }

        // line = strings.TrimSpace(line)
        line = Linefeed(line)
        file_xml_tmp_w.WriteString(line)

    }

    file_xml_tmp_r, err := os.Open(xml_tmp)
    check(err)
    defer file_xml_tmp_r.Close()

    buf_Tmp := bufio.NewReader(file_xml_tmp_r)
    file_oldMapper, err := os.Create(xml_output)
    check(err)
    defer file_oldMapper.Close()
    file_oldMapper.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    file_oldMapper.WriteString("<!DOCTYPE mapper PUBLIC \"-//mybatis.org//DTD Mapper 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-mapper.dtd\">\n")
    file_oldMapper.WriteString("<mapper namespace=\"SQLAudit." + xml_name + ".newMapper\">\n")
    file_oldMapper.WriteString("\n")

    flag1 := 0
    flag2 := 1
    flag3 := 0

    for {
        line, err := buf_Tmp.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println(xml_tmp + " read ok!")
                break
            } else {
                fmt.Println(xml_tmp+" read error!", err)
                common.LogIfError(err, "")
                return
            }
        }
        if len(strings.TrimSpace(line)) == 0 {
            continue
        } else if strings.Contains(strings.TrimSpace(line), "<!--") && !strings.Contains(strings.TrimSpace(line), "-->") {
            flag2 = 0
            continue
        } else if !strings.Contains(strings.TrimSpace(line), "<!--") && strings.Contains(strings.TrimSpace(line), "-->") {
            flag2 = 1
            continue
        } else if strings.Contains(line, "[") && flag2 == 1 {
            file_oldMapper.WriteString(line)
            flag3 = 1
        } else if strings.Contains(line, "]") && flag2 == 1 {
            file_oldMapper.WriteString(line)
            flag3 = 0
        } else if (strings.Contains(line, "<select ") || strings.Contains(line, "<select\n") || strings.Contains(line, "<select\t")) && flag2 == 1 {
            flag1 = 1
            line = Changeline(line)
            file_oldMapper.WriteString(line)
        } else if (strings.Contains(strings.TrimSpace(line), "</select>") || strings.Contains(strings.TrimSpace(line), "</select >") || strings.Contains(line, "</select\n") || strings.Contains(line, "</select\t")) && flag2 == 1 {
            flag1 = 0
            line = Changeline(line)
            file_oldMapper.WriteString(line + "\n")
        } else if strings.Contains(line, "<sql") && flag2 == 1 {
            flag1 = 1
            line = Changeline(line)
            file_oldMapper.WriteString(line)
        } else if strings.Contains(line, "</sql>") && flag2 == 1 {
            flag1 = 0
            line = Changeline(line)
            file_oldMapper.WriteString(line + "\n")
        } else if flag1 == 1 && flag2 == 1 && flag3 == 0 {
            line = Changeline(line)
            line = strings.Replace(line, "#", "\n#", -1)
            file_oldMapper.WriteString(line)
        } else if flag1 == 1 && flag2 == 1 && flag3 == 1 {
            line = strings.Replace(line, "#", "\n#", -1)
            file_oldMapper.WriteString(line)
        }
    }
    file_oldMapper.WriteString("</mapper>\n")
}

func Linefeed(line string) string {
    var s []int
    for i, v := range line {
        if string(v) == "\"" {
            s = append(s, i)
        }
    }
    oldline := line
    newline := line
    m := len(oldline)
    n := len(s)
    k := 0
    ii := 0
    for i := 0; i < m; i++ {
        flag1 := 0
        flag2 := 0
        if string(oldline[i]) == "<" {
            for j := 0; j < n; j += 2 {
                start := s[j]
                end := s[j+1]
                if i > start && i < end {
                    flag1 = 1
                }
            }
        }
        if string(oldline[i]) == ">" {
            for j := 0; j < n; j += 2 {
                start := s[j]
                end := s[j+1]
                if i > start && i < end {
                    flag2 = 1
                }
            }
        }
        if flag1 == 0 && string(oldline[i]) == "<" {
            ii = i + k
            newline = newline[0:ii] + "\n<" + newline[ii+1:]
            k = k + 1
        } else if flag2 == 0 && string(oldline[i]) == ">" {
            ii = i + k
            newline = newline[0:ii] + ">\n" + newline[ii+1:]
            k = k + 1
        }

    }
    return newline
}

func Changeline(line string) string {
    var s []int
    for i, v := range line {
        if string(v) == "\"" {
            s = append(s, i)
        }
    }
    oldline := line
    newline := line
    m := len(oldline)
    n := len(s)
    k := 0
    ii := 0
    for i := 0; i < m; i++ {
        flag1 := 0
        flag2 := 0
        if string(oldline[i]) == "<" {
            for j := 0; j < n; j += 2 {
                start := s[j]
                end := s[j+1]
                if i > start && i < end {
                    flag1 = 1
                }
            }
        }
        if string(oldline[i]) == ">" {
            for j := 0; j < n; j += 2 {
                start := s[j]
                end := s[j+1]
                if i > start && i < end {
                    flag2 = 1
                }
            }
        }
        if flag1 == 0 && string(oldline[i]) == "<" {
            ii = i + k
            newline = newline[0:ii] + "\n<<" + newline[ii+1:]
            k = k + 2
        } else if flag2 == 0 && string(oldline[i]) == ">" {
            ii = i + k
            newline = newline[0:ii] + ">>\n" + newline[ii+1:]
            k = k + 2
        }
    }
    return newline
}

// 格式化xml：先将<<和>>拼接为一行,然后将<<，>>替换回<，>
func Format(xml_output string, xml_output_formated string) {
    flag := 0
    lineTmp := ""
    file_r, err := os.Open(xml_output)
    check(err)
    defer file_r.Close()
    buf_Tmp := bufio.NewReader(file_r)
    file_w, err := os.Create(xml_output_formated)
    check(err)
    defer file_w.Close()

    for {
        line, err := buf_Tmp.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println(xml_output + " read ok!")
                break
            } else {
                fmt.Println(xml_output+" read error!", err)
                common.LogIfError(err, "")
                return
            }
        }

        if strings.Contains(line, "<<") && !strings.Contains(line, ">>") && flag == 0 {
            lineTmp = strings.TrimSpace(line)
            flag = 1
        } else if strings.Contains(line, ">>") && !strings.Contains(line, "<<") && flag == 1 {
            lineTmp = lineTmp + " " + strings.TrimSpace(line)
            flag = 0
            lineTmp = strings.Replace(lineTmp, "<<", "<", -1)
            lineTmp = strings.Replace(lineTmp, ">>", ">", -1)
            lineTmp = strings.Replace(lineTmp, "#", "\n#", -1)
            file_w.WriteString(lineTmp)
        } else if flag == 1 {
            lineTmp = lineTmp + " " + strings.TrimSpace(line)
        } else if flag == 0 {
            line = strings.Replace(line, "<<", "<", -1)
            line = strings.Replace(line, ">>", ">", -1)
            line = strings.Replace(line, "#", "\n#", -1)
            file_w.WriteString(line + "\n")
        }
    }
}

func isNotExist(slice []string, str string) bool {
    for _, i := range slice {
        if i == str {
            return false
        }
    }
    return true
}

func Remove_duplicates(column_list []string) []string {
    distinctList := []string{"alex"}
    for _, i := range column_list {
        if isNotExist(distinctList, i) {
            distinctList = append(distinctList, i)
        }
    }
    return distinctList[1:]
}

func Get_select(line string) []string {
    var s []int
    for i, v := range line {
        if string(v) == "\"" {
            s = append(s, i)
        }
    }
    idx1 := s[0] + 1
    idx2 := s[1]
    select_id := line[idx1:idx2]
    select_label := "<select id=\"" + select_id + "\">"
    var select_info []string
    select_info = append(select_info, select_id)
    select_info = append(select_info, select_label)
    return select_info
}

func Get_if(column_name string, line string) string {
    idx1 := strings.Index(line, "\"")
    tmp := line[0:idx1+1] + column_name + " !=null\">\n"
    return tmp
}

func Get_when(column_name string, line string) string {
    idx1 := strings.Index(line, "\"")
    tmp := line[0:idx1+1] + column_name + " !=null\">\n"
    return tmp
}

func Get_foreach(num int, select_info []string, line string) (string, string) {
    result := strings.Split(line, " ")
    // 获取“<”前面的空格并赋值给tmp，为了美观
    idx1 := strings.Index(line, "<")
    tmp := line[0:idx1]
    collection_column := ""
    for _, j := range result {
        if strings.Contains(j, "collection") {
            j = "collection=\"collection" + strconv.Itoa(num) + "\""
            collection_column = "collection" + strconv.Itoa(num)
        } else if strings.Contains(j, "item") {
            // 将item换成select_id+str(num),使其与foreach标签对中的变量相同
            j = "item=\"" + string(select_info[0]) + strconv.Itoa(num) + "\""
        }
        tmp = tmp + j + " "
    }
    // tmp = tmp + "\n"
    return collection_column, tmp
}

func Get_sql_dynamic(column_name string, line string) string {
    // 获取紧跟“#”号后面的“{”的索引idx2
    idx1 := strings.Index(line, "#")
    idx2 := 0
    idx3 := 0
    var s1 []int
    for i, v := range line {
        if string(v) == "\"" {
            s1 = append(s1, i)
        }
    }
    if len(s1) > 1 {
        for _, i := range s1 {
            if i > idx1 {
                idx2 = i
                break
            }
        }
    } else {
        idx2 = strings.Index(line, "{")
    }

    // 获取line中与上述“{”匹配的“}”的索引
    var s2 []int
    for i, v := range line {
        if string(v) == "\"" {
            s2 = append(s2, i)
        }
    }
    if len(s2) > 1 {
        for _, i := range s2 {
            if i > idx2 {
                idx3 = i
                break
            }
        }
    } else {
        idx3 = strings.Index(line, "}")
    }
    tmp := line[0:idx2+1] + column_name + line[idx3:]
    return tmp
}

func Deal_with_collection(column_list []string) []string {
    var column_list_final []string
    num := "0"
    num_len_reverse := 0
    for _, t := range column_list {
        if strings.Contains(t, "collection") {
            num = t[10:]
            num_len_reverse = -len(num)
            column_list_final = append(column_list_final, t)
        } else if t[len(t)+num_len_reverse:] == num {
            continue
        } else {
            column_list_final = append(column_list_final, t)
        }
    }
    return column_list_final
}

func Removeall(s []string, a string) []string {
    var result []string
    for _, v := range s {
        if v == a {
            continue
        } else {
            result = append(result, v)
        }
    }
    return result
}

func Remove(s []string, a string) []string {
    for i, v := range s {
        if v == a {
            s = append(s[:i], s[i+1:]...)
            return s
        }
    }
    return s
}

func isExist(s []string, a string) bool {
    for _, v := range s {
        if v == a {
            return true
        }
    }
    return false
}

func Get_fixed_list(column_list []string) []string {
    tmp := make([]string, len(column_list))
    copy(tmp, column_list)
    // 实际用于组合的动态字段会被记录两次
    tmp1 := Remove_duplicates(tmp)

    // fixed_list为固定列，即collection列和非null列
    fixed_list := make([]string, len(column_list))
    copy(fixed_list, column_list)

    for _, i := range tmp1 {
        tmp = Remove(tmp, i)
    }

    if len(tmp) > 0 {
        for _, m := range tmp {
            fixed_list = Removeall(fixed_list, m)
        }
    }
    return fixed_list
}

/*

定义组合方法start

*/

func combinations(s []string, m int) [][]string {
    n := len(s)
    indexs := zuheResult(n, m)
    result := findStrsByIndexs(s, indexs)
    return result
}

//组合算法(从nums中取出m个数)
func zuheResult(n int, m int) [][]int {
    if m < 1 || m > n {
        fmt.Println("Illegal argument. Param m must between 1 and len(nums).")
        return [][]int{}
    }

    //保存最终结果的数组，总数直接通过数学公式计算
    result := make([][]int, 0, mathZuhe(n, m))
    //保存每一个组合的索引的数组，1表示选中，0表示未选中
    indexs := make([]int, n)
    for i := 0; i < n; i++ {
        if i < m {
            indexs[i] = 1
        } else {
            indexs[i] = 0
        }
    }

    //第一个结果
    result = addTo(result, indexs)
    for {
        find := false
        //每次循环将第一次出现的 1 0 改为 0 1，同时将左侧的1移动到最左侧
        for i := 0; i < n-1; i++ {
            if indexs[i] == 1 && indexs[i+1] == 0 {
                find = true

                indexs[i], indexs[i+1] = 0, 1
                if i > 1 {
                    moveOneToLeft(indexs[:i])
                }
                result = addTo(result, indexs)

                break
            }
        }
        //本次循环没有找到 1 0 ，说明已经取到了最后一种情况
        if !find {
            break
        }
    }
    return result
}

//将ele复制后添加到arr中，返回新的数组
func addTo(arr [][]int, ele []int) [][]int {
    newEle := make([]int, len(ele))
    copy(newEle, ele)
    arr = append(arr, newEle)

    return arr
}

func moveOneToLeft(leftNums []int) {
    //计算有几个1
    sum := 0
    for i := 0; i < len(leftNums); i++ {
        if leftNums[i] == 1 {
            sum++
        }
    }

    //将前sum个改为1，之后的改为0
    for i := 0; i < len(leftNums); i++ {
        if i < sum {
            leftNums[i] = 1
        } else {
            leftNums[i] = 0
        }
    }
}

//根据索引号数组得到元素数组
func findStrsByIndexs(nums []string, indexs [][]int) [][]string {
    if len(indexs) == 0 {
        return [][]string{}
    }

    result := make([][]string, len(indexs))

    for i, v := range indexs {
        line := make([]string, 0)
        for j, v2 := range v {
            if v2 == 1 {
                line = append(line, nums[j])
            }
        }
        result[i] = line
    }

    return result
}

//数学方法计算排列数(从n中取m个数)
func mathPailie(n int, m int) int {
    return jieCheng(n) / jieCheng(n-m)
}

//数学方法计算组合数(从n中取m个数)
func mathZuhe(n int, m int) int {
    return jieCheng(n) / (jieCheng(n-m) * jieCheng(m))
}

//阶乘
func jieCheng(n int) int {
    result := 1
    for i := 2; i <= n; i++ {
        result *= i
    }
    return result
}

/*

end

*/

func Combine(column_list []string, fixed_list []string) [][]string {
    tmp1 := make([]string, len(column_list))
    copy(tmp1, column_list)
    for _, e := range fixed_list {
        tmp1 = Remove(tmp1, e)
    }
    // combined_list := make([][]string, 0)
    var combined_list [][]string
    if len(tmp1) == 0 {
        combined_list = append(combined_list, fixed_list)
        return combined_list
    }
    for i := 1; i < len(tmp1)+1; i++ {
        combined_list = append(combined_list, combinations(tmp1, i)...)
    }
    for i := 0; i < len(combined_list); i++ {
        combined_list[i] = append(combined_list[i], fixed_list...)
    }
    combined_list = append(combined_list, fixed_list)
    return combined_list
}

func CreateBean(select_info []string, bean_path string, column_list []string, package_name string) {
    file_bean, err := os.Create(bean_path)
    check(err)
    defer file_bean.Close()
    file_bean.WriteString(package_name + "\n")
    file_bean.WriteString("\n")
    for _, n := range column_list {
        if strings.Contains(n, "collection") {
            file_bean.WriteString("import java.util.List;\n")
            file_bean.WriteString("\n")
            break
        }
    }
    file_bean.WriteString("public class " + select_info[0] + "{\n")
    for _, n := range column_list {
        if strings.Contains(n, "collection") {
            file_bean.WriteString("private List " + n + ";\n")
        } else {
            file_bean.WriteString("private String " + n + ";\n")
        }
    }
    file_bean.WriteString("\n")
    for _, n := range column_list {
        if strings.Contains(n, "collection") {
            file_bean.WriteString("public List get" + strings.ToUpper(string(n[0])) + n[1:] + "() {\n")
            file_bean.WriteString("return " + n + ";\n")
            file_bean.WriteString("}\n")
            file_bean.WriteString("public void set" + strings.ToUpper(string(n[0])) + n[1:] + "(List " + n + "_P" + ")" + "{\n")
            file_bean.WriteString("this." + n + " = " + n + "_P;\n")
            file_bean.WriteString("}\n")
            file_bean.WriteString("\n")
        } else {
            file_bean.WriteString("public String get" + strings.ToUpper(string(n[0])) + n[1:] + "() {\n")
            file_bean.WriteString("return " + n + ";\n")
            file_bean.WriteString("}\n")
            file_bean.WriteString("public void set" + strings.ToUpper(string(n[0])) + n[1:] + "(String " + n + "_P" + "){\n")
            file_bean.WriteString("this." + n + " = " + n + "_P;\n")
            file_bean.WriteString("}\n")
            file_bean.WriteString("\n")
        }
    }
    file_bean.WriteString("}\n")
}

func CreateTestProcess(select_id string, combined_list [][]string, path_dir string, package_name string) {
    path := path_dir + "/" + "T_" + select_id + ".java"
    file_test, err := os.Create(path)
    check(err)
    defer file_test.Close()
    file_test.WriteString(package_name + "\n")
    file_test.WriteString("\n")
    for _, t := range combined_list[0] {
        if strings.Contains(t, "collection") {
            file_test.WriteString("import java.util.ArrayList;\n")
            file_test.WriteString("import java.util.List;\n")
            break
        }
    }
    file_test.WriteString("import org.apache.ibatis.session.SqlSession;\n")
    file_test.WriteString("\n")
    file_test.WriteString("public class " + "T_" + select_id + "{\n")
    file_test.WriteString("\n")
    file_test.WriteString("public static void main(String[] args) {\n")
    file_test.WriteString("\n")
    file_test.WriteString("SqlSession session = DBTools.getSession();\n")
    file_test.WriteString("newMapper mapper = session.getMapper(newMapper.class);\n")
    file_test.WriteString("\n")
    for _, t := range combined_list[0] {
        if strings.Contains(t, "collection") {
            file_test.WriteString("List<String> collection = new ArrayList<String>();\n")
            file_test.WriteString("collection.add(\"SQLAudit\");\n")
            file_test.WriteString("\n")
            break
        }
    }
    x := 0
    for _, y := range combined_list {
        file_test.WriteString(select_id + " " + select_id + "_P" + strconv.Itoa(x) + " = " + "new " + select_id + "();\n")
        for _, z := range y {
            if strings.Contains(z, "collection") {
                file_test.WriteString(select_id + "_P" + strconv.Itoa(x) + "." + "set" + strings.ToUpper(string(z[0])) + z[1:] + "(" + "collection" + ");\n")
            } else {
                file_test.WriteString(select_id + "_P" + strconv.Itoa(x) + "." + "set" + strings.ToUpper(string(z[0])) + z[1:] + "(" + "\"SQLAudit\"" + ");\n")
            }
        }
        file_test.WriteString(select_id + "(" + select_id + "_P" + strconv.Itoa(x) + ", session, mapper);\n")
        file_test.WriteString("\n")
        x++
    }
    file_test.WriteString("session.close();\n")
    file_test.WriteString("\n")
    file_test.WriteString("}\n")
    file_test.WriteString("\n")
    file_test.WriteString("private static void " + select_id + "(" + select_id + " " + select_id + "_P, SqlSession session, newMapper mapper) {\n")
    file_test.WriteString("\n")
    file_test.WriteString("try {\n")
    file_test.WriteString("mapper." + select_id + "(" + select_id + "_P" + ");\n")
    file_test.WriteString("session.commit();\n")
    file_test.WriteString("} " + "catch (Exception e) {\n")
    file_test.WriteString("session.rollback();\n")
    file_test.WriteString("}\n")
    file_test.WriteString("}\n")
    file_test.WriteString("}\n")
}

func CreateConf(path_dir string, xml_name string, test_db string, username string, password string) {
    file_conf, err := os.Create(path_dir + "/conf.xml")
    check(err)
    defer file_conf.Close()
    file_conf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    file_conf.WriteString("<!DOCTYPE configuration PUBLIC \"-//mybatis.org//DTD Config 3.0//EN\" \"http://mybatis.org/dtd/mybatis-3-config.dtd\">\n")
    file_conf.WriteString("\n")
    file_conf.WriteString("<configuration>\n")
    file_conf.WriteString("   <environments default=\"development\">\n")
    file_conf.WriteString("           <environment id=\"development\">\n")
    file_conf.WriteString("                   <transactionManager type=\"JDBC\" />\n")
    file_conf.WriteString("                   <dataSource type=\"POOLED\">\n")
    file_conf.WriteString("                           <property name=\"driver\" value=\"com.mysql.jdbc.Driver\" />\n")
    file_conf.WriteString("                           <property name=\"url\" value=\"jdbc:mysql://" + test_db + "?characterEncoding=UTF-8\"/>\n")
    file_conf.WriteString("                           <property name=\"username\" value=\"" + username + "\"/>\n")
    file_conf.WriteString("                           <property name=\"password\" value=\"" + password + "\"/>\n")
    file_conf.WriteString("                   </dataSource>\n")
    file_conf.WriteString("           </environment>\n")
    file_conf.WriteString("   </environments>\n")
    file_conf.WriteString("\n")
    file_conf.WriteString("   <mappers>\n")
    file_conf.WriteString("           <mapper class=\"" + "SQLAudit." + xml_name + ".newMapper\" />\n")
    file_conf.WriteString("   </mappers>\n")
    file_conf.WriteString("\n")
    file_conf.WriteString("</configuration>")
}

func Create_log4j_properties(path_dir string, path_log string, xml_name string) {
    file_log4j_properties, err := os.Create(path_dir + "/log4j.properties")
    check(err)
    defer file_log4j_properties.Close()
    file_log4j_properties.WriteString("### set log levels ###\n")
    file_log4j_properties.WriteString("log4j.rootLogger = INFO, stdConsole, stdFile\n")
    file_log4j_properties.WriteString("log4j.logger.com.paic.dbaudit = DEBUG\n")
    file_log4j_properties.WriteString("\n")
    file_log4j_properties.WriteString("### set trace###\n")
    file_log4j_properties.WriteString("log4j.logger.SQLAudit." + xml_name + " = TRACE\n")
    file_log4j_properties.WriteString("\n")
    file_log4j_properties.WriteString("### output to the console ###\n")
    file_log4j_properties.WriteString("log4j.appender.stdConsole = org.apache.log4j.ConsoleAppender\n")
    file_log4j_properties.WriteString("log4j.appender.stdConsole.layout = org.apache.log4j.PatternLayout\n")
    file_log4j_properties.WriteString("log4j.appender.stdConsole.layout.ConversionPattern = %-d{yyyy-MM-dd HH:mm:ss} [%c.%M:%L]-[%p] %m%n\n")
    file_log4j_properties.WriteString("\n")
    file_log4j_properties.WriteString("### output to the log file ###\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile = org.apache.log4j.DailyRollingFileAppender\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.layout = org.apache.log4j.PatternLayout\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.File = " + path_log + "/" + xml_name + ".log\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.DatePattern='.'yyyy-MM-dd\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.Append = true\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.Threshold = debug\n")
    file_log4j_properties.WriteString("log4j.appender.stdFile.layout.ConversionPattern = %-d{yyyy-MM-dd HH:mm:ss} [%c.%M:%L]-[%p] %m%n\n")
}

func Create_DBTools(path_dir string, package_name string) {
    file_DBTools, err := os.Create(path_dir + "/DBTools.java")
    check(err)
    defer file_DBTools.Close()
    file_DBTools.WriteString(package_name + "\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("import java.io.Reader;\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("import org.apache.ibatis.io.Resources;\n")
    file_DBTools.WriteString("import org.apache.ibatis.session.SqlSession;\n")
    file_DBTools.WriteString("import org.apache.ibatis.session.SqlSessionFactory;\n")
    file_DBTools.WriteString("import org.apache.ibatis.session.SqlSessionFactoryBuilder;\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("public class DBTools {\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("    public static SqlSessionFactory sessionFactory;\n")
    file_DBTools.WriteString("    static{\n")
    file_DBTools.WriteString("        try {\n")
    file_DBTools.WriteString("            Reader reader = Resources.getResourceAsReader(\"conf.xml\");\n")
    file_DBTools.WriteString("            sessionFactory = new SqlSessionFactoryBuilder().build(reader);\n")
    file_DBTools.WriteString("        } catch (Exception e) {\n")
    file_DBTools.WriteString("            e.printStackTrace();\n")
    file_DBTools.WriteString("        }\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("    }\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("    public static SqlSession getSession(){\n")
    file_DBTools.WriteString("        return sessionFactory.openSession();\n")
    file_DBTools.WriteString("    }\n")
    file_DBTools.WriteString("\n")
    file_DBTools.WriteString("}")
}

func CreateInterface(id_list []string, path_dir string, package_name string) {
    file_Interface, err := os.Create(path_dir + "/newMapper.java")
    check(err)
    defer file_Interface.Close()
    file_Interface.WriteString(package_name + "\n")
    file_Interface.WriteString("\n")
    file_Interface.WriteString("public interface newMapper {\n")
    for _, m := range id_list {
        file_Interface.WriteString("void " + m + "(" + m + " " + m + "_P" + ");\n")
    }
    file_Interface.WriteString("}\n")
}

func CreateSh(path_dir string, path_lib string, path_sh string, id_list []string, xml_name string) {
    file_sh, err := os.Create(path_sh + "/" + xml_name + ".sh")
    check(err)
    defer file_sh.Close()
    file_sh.WriteString("\n")
    file_sh.WriteString("echo \"Compiling java files...\"\n")
    file_sh.WriteString("javac -d " + path_dir + " -cp " + path_lib + "/mybatis-3.2.8.jar " + path_dir + "/DBTools.java\n")
    for _, a := range id_list {
        file_sh.WriteString("javac -d " + path_dir + " " + path_dir + "/" + a + ".java\n")
    }
    file_sh.WriteString("javac -d " + path_dir + " -cp " + path_dir + " " + path_dir + "/newMapper.java\n")
    for _, b := range id_list {
        file_sh.WriteString("javac -d " + path_dir + " -cp " + path_lib + "/mybatis-3.2.8.jar:" + path_dir + " " + path_dir + "/" + "T_" + b + ".java\n")
    }
    file_sh.WriteString("\n")
    file_sh.WriteString("echo \"Copying newMapper.xml to directory where .class files have been...\"\n")
    file_sh.WriteString("cp " + path_dir + "/newMapper.xml " + path_dir + "/SQLAudit/" + xml_name + "\n")
    file_sh.WriteString("\n")
    file_sh.WriteString("echo \"executing .class files and writing sql to " + xml_name + ".log...\"\n")
    file_sh.WriteString("cd " + path_dir + "\n")
    for _, c := range id_list {
        file_sh.WriteString("java -cp " + path_lib + "/*:. SQLAudit." + xml_name + "." + "T_" + c + "\n")
    }
}

func ParseMapper(xml_name string, path_root string, path_file string, path_sh string, path_lib string, path_log string, test_db string, username string, password string) [][]string {
    // 创建${xml_name}格式的目录，用于存放每次解析生成的文件
    path_dir := path_root + "/tmp/" + xml_name
    err := Makedir(path_dir)
    if err == nil {
        fmt.Println("Dir '" + path_dir + "' was created/rebuilt successfully!")
    } else {
        common.LogIfError(err, "")
        os.Exit(1)
    }

    // 创建配置文件
    CreateConf(path_dir, xml_name, test_db, username, password)

    // 创建log4j.properties
    Create_log4j_properties(path_dir, path_log, xml_name)

    // 创建DBTools.java
    package_name := "package SQLAudit." + xml_name + ";"
    Create_DBTools(path_dir, package_name)

    // 获取select语句，将<，>替换为<<，>>并写入oldMapper.xml
    xml_input := path_file + "/" + xml_name + ".xml"
    xml_output := path_dir + "/oldMapper.xml"
    Filter(path_dir, xml_input, xml_output, xml_name)

    // 进一步格式化oldMapper.xml，将<<，>>替换回<，>并写入oldMapperFormated.xml
    xml_output_formated := path_dir + "/oldMapperFormated.xml"
    Format(xml_output, xml_output_formated)

    // 根据oldMapper.xml生成精简版newMapper.xml
    file_oldMapper, err := os.Open(xml_output_formated)
    check(err)
    defer file_oldMapper.Close()
    buf_Tmp := bufio.NewReader(file_oldMapper)
    file_newMapper, err := os.Create(path_dir + "/newMapper.xml")
    check(err)
    defer file_newMapper.Close()
    // 定义变量
    var id_list []string
    var id_too_many_list []string
    var column_list []string
    var select_info []string
    var column_name string
    var if_null_sql string
    var if_no_null_sql string
    var when_null_sql string
    var when_no_null_sql string
    var sql_dynamic string
    var fixed_list []string
    var bean_path string
    var id_sql [][]string
    var i int

    // 开始构建newMapper.xml
    for {
        line, err := buf_Tmp.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println(xml_output_formated + " read ok!")
                break
            } else {
                fmt.Println(xml_output_formated+" read error!", err)
                common.LogIfError(err, "")
                return id_sql
            }
        }
        if len(strings.TrimSpace(line)) == 0 {
            continue
        } else if strings.Contains(line, "<select") {
            i = 0
            column_list = []string{}
            select_info = Get_select(line)
            file_newMapper.WriteString("\n\n" + select_info[1] + "\n")
        } else if strings.Contains(line, "--") {
            continue
        } else if strings.Contains(line, "<if") {
            if strings.Contains(line, "null") {
                column_name = select_info[0] + strconv.Itoa(i)
                column_list = append(column_list, column_name)
                if_null_sql = Get_if(column_name, line)
                file_newMapper.WriteString(if_null_sql)
            } else {
                column_name = select_info[0] + strconv.Itoa(i)
                // 记录两次，把该列当做动态字段（动态字段会被记录两次）
                column_list = append(column_list, column_name)
                column_list = append(column_list, column_name)
                if_no_null_sql = Get_if(column_name, line)
                file_newMapper.WriteString(if_no_null_sql)
                i++
            }
        } else if strings.Contains(line, "<when") {
            if strings.Contains(line, "null") {
                column_name = select_info[0] + strconv.Itoa(i)
                column_list = append(column_list, column_name)
                when_null_sql = Get_when(column_name, line)
                file_newMapper.WriteString(when_null_sql)
            } else {
                column_name = select_info[0] + strconv.Itoa(i)
                // 记录两次，把该列当做动态字段（动态字段会被记录两次）
                column_list = append(column_list, column_name)
                column_list = append(column_list, column_name)
                when_no_null_sql = Get_if(column_name, line)
                file_newMapper.WriteString(when_no_null_sql)
                i++
            }
        } else if strings.Contains(line, "<foreach") {
            foreach_column, foreach_line := Get_foreach(i, select_info, line)
            column_list = append(column_list, foreach_column)
            file_newMapper.WriteString(foreach_line)
        } else if strings.Contains(line, "#") {
            column_name = select_info[0] + strconv.Itoa(i)
            column_list = append(column_list, column_name)
            sql_dynamic = Get_sql_dynamic(column_name, line)
            file_newMapper.WriteString(sql_dynamic)
            i++
        } else if strings.Contains(line, "</select") {
            file_newMapper.WriteString(line + "\n\n")
            id_list = append(id_list, select_info[0])
            fixed_list = Get_fixed_list(column_list)
            column_list = Remove_duplicates(column_list)

            for _, s := range column_list {
                if strings.Contains(s, "collection") {
                    column_list = Deal_with_collection(column_list)
                    break
                }
            }

            for _, v := range fixed_list {
                if strings.Contains(v, "collection") {
                    fixed_list = Deal_with_collection(fixed_list)
                    break
                }
            }
            n := len(column_list) - len(fixed_list)
            if n <= 8 {
                combined_list := Combine(column_list, fixed_list)
                bean_path = path_dir + "/" + select_info[0] + ".java"
                CreateBean(select_info, bean_path, column_list, package_name)
                CreateTestProcess(select_info[0], combined_list, path_dir, package_name)
            } else {
                id_list = Remove(id_list, select_info[0])
                id_too_many_list = append(id_too_many_list, select_info[0])

            }
        } else {
            file_newMapper.WriteString(line)
        }
    }

    // 创建接口文件
    CreateInterface(id_list, path_dir, package_name)

    // 创建Sh文件
    CreateSh(path_dir, path_lib, path_sh, id_list, xml_name)

    // 清理之前生成的sql日志
    err = os.Remove(path_log + "/" + xml_name + ".log")
    if err == nil {
        fmt.Println("Ex-logfile was removed successfully!")
    }

    // 执行shell调起Mybatis工程
    ExecMybatisProject(path_sh, xml_name)

    // 解析Mybatis日志获取select_id和sql文本
    logPath := path_log + "/" + xml_name + ".log"
    id_sql = Get_SQL(logPath)
    return id_sql
}

func ExecMybatisProject(path_sh string, xml_name string) {
    cmdStr := "sh " + path_sh + "/" + xml_name + ".sh"
    cmd := exec.Command("/bin/bash", "-c", cmdStr)
    err := cmd.Run()
    if err != nil {
        common.LogIfError(err, "")
        os.Exit(1)
    } else {
        fmt.Println("Java files were compiled and executed successfully! ")
    }
}

func Get_SQL(logPath string) [][]string {
    sqlLog, err := os.Open(logPath)
    check(err)
    buf_Tmp := bufio.NewReader(sqlLog)
    defer sqlLog.Close()

    var id_sql [][]string
    var select_id string
    var select_sql string
    var result []string
    for {
        line, err := buf_Tmp.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println(logPath + " read ok!")
                break
            } else {
                fmt.Println(logPath+" read error!", err)
                common.LogIfError(err, "")
                return id_sql
            }
        }
        if strings.Contains(line, "==>  Preparing:") {
            result = strings.Split(line, " ")
            // result的第3个元素按“.”分割后的倒数第2个元素为select id
            select_id = strings.Split(result[2], ".")[len(strings.Split(result[2], "."))-2]
            select_sql = ""
            for i := 6; i < len(result); i++ {
                // 去掉开头的空格
                if len(select_sql) == 0 {
                    select_sql = result[i]
                } else {
                    select_sql = select_sql + " " + result[i]
                }
            }
            id_sql = append(id_sql, []string{select_id, select_sql})
        }
    }
    return id_sql
}

func WriteSQL2File(id_sql [][]string, path_sql string, xml_name string, db *sql.DB, dbName string) []string {

    err := Makedir(path_sql + "/" + xml_name)
    if err == nil {
        fmt.Println("Dir '" + path_sql + "/" + xml_name + "' was created/rebuilt successfully!")
    } else {
        common.LogIfError(err, "")
        os.Exit(1)
    }

    newQuery := ""
    select_id := []string{}
    for _, i := range id_sql {
        select_id = append(select_id, i[0])
    }
    select_id = Remove_duplicates(select_id)

    for _, k := range select_id {
        sqltext, err := os.Create(path_sql + "/" + xml_name + "/" + k + ".sql")
        check(err)
        defer sqltext.Close()

        for _, i := range id_sql {
            if i[0] == k {
                sqltext.WriteString("#" + strings.TrimSpace(i[1]) + ";\n")
                newQuery = ReplaceQuestionMark(i[1], db, dbName)
                sqltext.WriteString(" " + strings.TrimSpace(newQuery) + ";\n")
            }
        }

    }
    return select_id
}

func MapperAudit(test_dsn string, username string, password string, select_id []string, path_sql string, path_audit string, xml_name string, report_type string) {
    dsn4soar := "\"" + username + ":" + password + "@" + test_dsn + "\""
    // 在audit下创建${xml_name}目录
    err := Makedir(path_audit + "/" + xml_name)
    if err == nil {
        fmt.Println("Dir '" + path_audit + "/" + xml_name + "' was created/rebuilt successfully!")
    } else {
        common.LogIfError(err, "")
        os.Exit(1)
    }
    for _, i := range select_id {
        cmdStr := "soar -test-dsn=" + dsn4soar + " -allow-online-as-test -query " + path_sql + "/" + xml_name + "/" + i + ".sql -report-type " + report_type + " >" + path_audit + "/" + xml_name + "/" + i + "_audit." + report_type
        cmd := exec.Command("/bin/bash", "-c", cmdStr)
        err := cmd.Run()
        if err != nil {
            common.LogIfError(err, "")
            os.Exit(1)
        } else {
            fmt.Println("xml file was sucessfully audited, you can find the audit file '" + i + "_audit." + report_type + "' in directory " + path_audit + "/" + xml_name + ".")
        }
    }
}

