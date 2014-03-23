package dao


type Column struct {
    Name    string
    Type    string
    Length  int
    Primary bool
}


type Table struct {
    DB string
    Name string
    Primary []int
    Cols []int
}


type Dao struct {
    Name string
    Table string
    Cols []int
}

var (
    columns []*Column
    tables map[string]*Table
    daopool map[string]*Dao
)

func load
