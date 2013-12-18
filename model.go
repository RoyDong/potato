package potato

import (
    "fmt"
    "strings"
    "database/sql"
    _"github.com/go-sql-driver/mysql"
)

type Model struct {
    Table string
    Columns []string
}


func (m *Model) SqlColumnsPart(Columns []string) string {
    return "`" + strings.Join(Columns, "`,`") + "`"
}

func (m *Model) CreateFindStmt(query map[string]interface{}, order string, limit ...int) (string, []interface{}) {
    var stmt, where string
    var length = len(query)
    var values = make([]interface{}, 0, length)

    if length > 0 {
        conditions := make([]string, 0, length)
        for k, v := range query {
            i := strings.Split(k, " ")
            c := i[0]
            o := "="
            if len(i) > 1 {
                o = i[1]
            }

            conditions = append(conditions, fmt.Sprintf("`%s` %s ?", c, o))
            values = append(values, v)
        }

        where = strings.Join(conditions, " AND ")
    }

    if len(limit) == 1 {
        stmt = fmt.Sprintf("SELECT %s FROM `%s` %s ORDER BY %s LIMIT %d",
                m.SqlColumnsPart(m.Columns), m.Table, where, order, limit[0])
    } else if len(limit) == 2 {
        stmt = fmt.Sprintf("SELECT %s FROM `%s` %s ORDER BY %s LIMIT %d, %d",
                m.SqlColumnsPart(m.Columns), m.Table, where, order, limit[0], limit[1])
    }

    return stmt, values
}

func (m *Model) Find(query map[string]interface{}, order string, limit ...int) (*sql.Rows, error) {
    stmt, values := m.CreateFindStmt(query, order, limit...)
    rows, e := D.Query(stmt, values...)
    return rows, e
}

func (m *Model) FindOne(query map[string]interface{}, order string) *sql.Row {
    stmt, values := m.CreateFindStmt(query, order, 1)
    return D.QueryRow(stmt, values...)
}
