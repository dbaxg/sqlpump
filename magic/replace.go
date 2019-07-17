package magic

import (
    "database/sql"
    "fmt"
    "strconv"
    "strings"

    "github.com/myaudit/common"
    "vitess.io/vitess/go/vt/sqlparser"
)

func GetInfo(node sqlparser.SQLNode) ([][]string, []string, [][]int, error) {
    var columns [][]string
    var tables []string
    var limits [][]int
    // var havings [][]string
    conditions, err := FindAllCondition(node)
    if err != nil {
        return columns, tables, limits, err
    }
    for _, cond := range conditions {
        switch node := cond.(type) {
        case *sqlparser.ComparisonExpr:
            // TODO 考虑 ？= id这种情况，参考replace.go中的隐式转换函数
            // 考虑 from db1.test1, sakila.film where db1.test1.id = sakila.film.id
            // from db1.test1, sakila.film a where db1.test1.id = a.id 写个函数根据a去找到 sakila.film
            // TODO 函数中的问号还没解决
            // select count(*) FROM film where id = to_number( ?,'99999999') and language_id = ?
            // select count(*) FROM film where id = to_number( ?,'99999999') and language_id = 1

            // 获取 condition 左侧的信息
            switch nLeft := node.Left.(type) {
            case *sqlparser.SQLVal, sqlparser.ValTuple:
                err := sqlparser.Walk(func(nLeft sqlparser.SQLNode) (kontinue bool, err error) {
                    switch val := nLeft.(type) {
                    case *sqlparser.SQLVal:
                        if val.Type == 5 && val.Val[0] == 0x3a && val.Val[1] == 0x76 {
                            switch col := node.Right.(type) {
                            case *sqlparser.ColName:
                                var column []string
                                column = append(column, col.Name.String())
                                column = append(column, string(val.Val[2]))
                                columns = append(columns, column)
                            }
                        }
                    }
                    return true, nil
                }, nLeft)
                if err != nil {
                    return columns, tables, limits, err
                }
                // fmt.Println(err)
            }

            // 获取 condition 右侧的信息
            switch nRight := node.Right.(type) {
            case *sqlparser.SQLVal, sqlparser.ValTuple:
                err := sqlparser.Walk(func(nRight sqlparser.SQLNode) (kontinue bool, err error) {
                    switch val := nRight.(type) {
                    case *sqlparser.SQLVal:
                        if val.Type == 5 && val.Val[0] == 0x3a && val.Val[1] == 0x76 {
                            switch col := node.Left.(type) {
                            case *sqlparser.ColName:
                                var column []string
                                column = append(column, col.Name.String())
                                column = append(column, string(val.Val[2]))
                                columns = append(columns, column)
                            }
                        }
                    }
                    return true, nil
                }, nRight)
                if err != nil {
                    return columns, tables, limits, err
                }
                // fmt.Println(err)
            }
        case *sqlparser.RangeCond:
            switch fNode := node.From.(type) {
            case *sqlparser.SQLVal:
                if fNode.Type == 5 && fNode.Val[0] == 0x3a && fNode.Val[1] == 0x76 {
                    switch col := node.Left.(type) {
                    case *sqlparser.ColName:
                        var column []string
                        column = append(column, col.Name.String())
                        column = append(column, string(fNode.Val[2]))
                        columns = append(columns, column)
                    }
                }
            }

            switch tNode := node.To.(type) {
            case *sqlparser.SQLVal:
                if tNode.Type == 5 && tNode.Val[0] == 0x3a && tNode.Val[1] == 0x76 {
                    switch col := node.Left.(type) {
                    case *sqlparser.ColName:
                        var column []string
                        column = append(column, col.Name.String())
                        column = append(column, string(tNode.Val[2]))
                        columns = append(columns, column)
                    }
                }
            }
        case *sqlparser.IsExpr:
            // TODO
        }
    }

    tableNodes, err := FindAllTable(node)
    if err != nil {
        return columns, tables, limits, err
    }
    for _, tab := range tableNodes {
        switch node := tab.(type) {
        case *sqlparser.AliasedTableExpr:
            switch tabName := node.Expr.(type) {
            case sqlparser.TableName:
                tables = append(tables, tabName.Name.String())
            }
        }
    }

    limitNodes, err := FindLimit(node)
    if err != nil {
        return columns, tables, limits, err
    }

    for _, lim := range limitNodes {
        switch node := lim.(type) {
        case *sqlparser.Limit:
            var limit []int
            if node != nil {
                switch offVal := node.Offset.(type) {
                case *sqlparser.SQLVal:
                    if offVal != nil {
                        if offVal.Type == 5 && offVal.Val[0] == 0x3a && offVal.Val[1] == 0x76 {
                            position, _ := strconv.Atoi(string(offVal.Val[2]))
                            limit = append(limit, position)
                        } else {
                            limit = append(limit, -1)
                        }
                    } else {
                        limit = append(limit, -1)
                    }
                }

                switch rowVal := node.Rowcount.(type) {
                case *sqlparser.SQLVal:
                    if rowVal != nil {
                        if rowVal.Type == 5 && rowVal.Val[0] == 0x3a && rowVal.Val[1] == 0x76 {
                            position, _ := strconv.Atoi(string(rowVal.Val[2]))
                            limit = append(limit, position)
                        } else {
                            limit = append(limit, -1)
                        }
                    } else {
                        limit = append(limit, -1)
                    }
                }
            } else {
                limit = append(limit, -1)
                limit = append(limit, -1)
            }
            limits = append(limits, limit)
        }
    }

    return columns, tables, limits, nil
}

func FindAllCondition(node sqlparser.SQLNode) ([]interface{}, error) {
    var conditions []interface{}
    err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
        switch node := node.(type) {
        case *sqlparser.ComparisonExpr, *sqlparser.RangeCond, *sqlparser.IsExpr:
            conditions = append(conditions, node)
        }
        return true, nil
    }, node)
    return conditions, err
}

func FindAllTable(node sqlparser.SQLNode) ([]interface{}, error) {
    var tableNodes []interface{}
    err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
        switch node := node.(type) {
        case *sqlparser.AliasedTableExpr:
            tableNodes = append(tableNodes, node)
        }
        return true, nil
    }, node)
    // fmt.Println(err)
    return tableNodes, err
}

func FindLimit(node sqlparser.SQLNode) ([]interface{}, error) {
    var limitNodes []interface{}
    err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
        switch node := node.(type) {
        case *sqlparser.Limit:
            limitNodes = append(limitNodes, node)
        }
        return true, nil
    }, node)
    // fmt.Println(err)
    return limitNodes, err
}

func ReplaceQuestionMark(query string, db *sql.DB, dbName string) string {
    stmt, err := sqlparser.Parse(query)
    if err != nil {
        common.LogIfError(err, "parse failed: "+query)
        return query
    }

    columns, tables, _, err := GetInfo(stmt)
    if err != nil {
        common.LogIfError(err, "get info failed: "+query)
        return query
    }

    var column_name string
    var data_type string

    column_type := make(map[string]string)

    for _, tab := range tables {
        rows, _ := db.Query("select column_name, data_type from columns where table_name = ? and TABLE_SCHEMA = '"+dbName+"'", tab)
        for rows.Next() {
            if err := rows.Scan(&column_name, &data_type); err == nil {
                if _, ok := column_type[column_name]; !ok {
                    column_type[column_name] = data_type
                }
            } else {
                fmt.Println(err)
            }
        }
        rows.Close()
    }

    i := 0
    for _, col := range columns {

        if column, ok := column_type[col[0]]; ok {
            i++
            delete(column_type, col[0])
            // fmt.Println("here is", query)
            question_index := GetIndex(query, "?")
            if question_index != nil {
                // fmt.Println("this is :", question_index)
                idx, _ := strconv.Atoi(col[1])
                // fmt.Println("idx is :", idx)
                index := question_index[idx-i]
                // fmt.Println("index is :", index)
                if strings.Contains(column, "char") || strings.Contains(column, "text") || strings.Contains(column, "lob") {
                    query = query[:index] + "'SQLAudit'" + query[index+1:]
                } else if strings.Contains(column, "int") {
                    query = query[:index] + "1" + query[index+1:]
                } else if strings.Contains(column, "decimal") || strings.Contains(column, "float") || strings.Contains(column, "double") || strings.Contains(column, "numeric") {
                    query = query[:index] + "1.1" + query[index+1:]
                } else if strings.Contains(column, "timestamp") || strings.Contains(column, "datatime") {
                    query = query[:index] + "'2019-01-01 00:00:00'" + query[index+1:]
                } else if strings.Contains(column, "date") {
                    query = query[:index] + "'2019-01-01'" + query[index+1:]
                } else if strings.Contains(column, "time") {
                    query = query[:index] + "'00:00:00'" + query[index+1:]
                } else if strings.Contains(column, "year") {
                    query = query[:index] + "'2019'" + query[index+1:]
                } else if strings.Contains(column, "enum") {
                    query = query[:index] + "'SQLAudit'" + query[index+1:]
                } else {
                    query = query[:index] + "none" + query[index+1:]
                }

            }
        }
    }

    stmt, err = sqlparser.Parse(query)
    if err != nil {
        common.LogIfError(err, "parse failed: "+query)
        return query
    }

    _, _, limits, err := GetInfo(stmt)
    if err != nil {
        common.LogIfError(err, "get info failed: "+query)
        return query
    }

    // 替换limit中的？
    for _, ques := range limits {
        question_index := GetIndex(query, "?")
        if ques[0] != -1 {
            index := question_index[ques[0]-1]
            query = query[:index] + "1" + query[index+1:]
        }
        if ques[1] != -1 {
            index := question_index[ques[1]-1]
            query = query[:index] + "1" + query[index+1:]
        }
    }

    // TODO group by和order by中的？替换。
    // 放弃，因为group by和order by的对象是列而不是值，而列我们是不知道的，所以没法替换

    // 打印信息
    // fmt.Println("columns:", columns)
    // fmt.Println("tables:", tables)
    // fmt.Println("limits:", limits)
    // fmt.Println("new query:", query)
    return query
}

func GetIndex(str string, subStr string) []int {
    var result []int
    length := len(subStr)
    tmpStr := ""
    for i := 0; i < length; i++ {
        tmpStr = tmpStr + "1"
    }
    for {
        if !strings.Contains(str, subStr) {
            break
        }
        index := strings.Index(str, subStr)
        result = append(result, index)
        str = str[:index] + tmpStr + str[index+1:]
    }
    return result
    // index := strings.Contains(str,subStr)
}

