[![Build Status](https://travis-ci.org/MathieuNls/go-sql-wrapper.png)](https://travis-ci.org/MathieuNls/go-sql-wrapper)
[![GoDoc](https://godoc.org/github.com/MathieuNls/go-sql-wrapper?status.png)](https://godoc.org/github.com/MathieuNls/go-sql-wrapper)
[![codecov](https://codecov.io/gh/MathieuNls/go-sql-wrapper/branch/master/graph/badge.svg)](https://codecov.io/gh/MathieuNls/go-sql-wrapper)


# go-sql-wrapper

go-sql-wapper is not an orm but, as its name suggests, a sql-wrapper. It allows you do to:

- [insert](#Insert)
- [select](#Select)
- [delete](#Delete)
- update (WIP)

in a type safe, struct directed way. 

## Insert

```go
type MyStuct struct {
    dbId             int     `db:id`
	ExportedInted    int     `db:"aaa"`
	ExportedString   string  `db:"bbb"`
	ExportedFloat64  float64 `db:"ccc"`
}

func main() {
    model, err := NewSQLModel("MyTable", []string{"root:root@tcp(127.0.0.1:3306)/mydb", new(MySQLCnxOpenner))

    myStruct := new(MyStruct)
    myStruct.ExportedInted = 2
    myStruct.ExportedString = "string"
    myStruct.ExportedFloat64 = 1.0

    model.insert(myStruct); 
    //Executes
    //INSERT INTO MyTable (aaa, bbb, ccc) VALUES (2, 'string', 1.0);
}
```

## Select

```go
func main() {

    result := model.select("aaa, bbb").Find("1")
    //Produces Select aaa,bbb where id = 1
    fmt.Println(result)
    //Prints MyStruct{1, 2, string, 0.0 }
    //Fetched MyStruct from the db, ExportedFloat64 is not populated as ccc wasn't requested
}
```

## Delete
```go
func main() {

    result := model.select("aaa, bbb").Find("1")
    //Produces Select aaa,bbb where id = 1
    fmt.Println(result)
    //Prints MyStruct{1, 2, string, 0.0 }
    //Fetched MyStruct from the db, ExportedFloat64 is not populated as ccc wasn't requested
}
```


# Why ?

Why would you want an SQL wrapper in Go ? Well, don't you like to have enforced types and the additional safety that come with it ? Yes, then, this 
 
```go
stmtOut, err := model.db.Prepare("Select * from myTable")

if err != nil {
    panic(err)
}
defer stmtOut.Close()
rows, err := stmtOut.Query()

columns, err := rows.Columns()
if err != nil {
    panic(err)
}

// Make a slice for the values
values := make([]sql.RawBytes, len(columns))

// rows.Scan wants '[]interface{}' as an argument, so we must copy the
// references into such a slice
// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
scanArgs := make([]interface{}, len(values))
for i := range values {
    scanArgs[i] = &values[i]
}

// Fetch rows
for rows.Next() {
    // get RawBytes from data

    err = rows.Scan(scanArgs...)

    if err != nil {
        panic(err)
    }

    myInt := scanArgs[0]
    myString := scanArgs[1]
    //...
}
```

should not be so pleasing...

# Where do we Go from there

- [ ] documentation
- [ ] set of examples
- [x] Mysql
- [ ] Oracle
- [ ] PostgresSQL