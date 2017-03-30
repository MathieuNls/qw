[![Build Status](https://travis-ci.org/MathieuNls/go-sql-wrapper.png)](https://travis-ci.org/MathieuNls/go-sql-wrapper)
[![GoDoc](https://godoc.org/github.com/MathieuNls/go-sql-wrapper?status.png)](https://godoc.org/github.com/MathieuNls/go-sql-wrapper)
[![codecov](https://codecov.io/gh/MathieuNls/go-sql-wrapper/branch/master/graph/badge.svg)](https://codecov.io/gh/MathieuNls/go-sql-wrapper)


# go-sql-wrapper

go-sql-wapper is not an orm but, as its name suggests, a sql-wrapper. It allows you do to:

- [insert](#insert)
- [select](#select)
- [delete](#delete)
- update (WIP)

in a type safe, struct directed way. 

The `SQLModel` struct that provide most of the heavy lifting is a _[fluent interface](https://www.wikiwand.com/en/Fluent_interface)_; meaning that call can be linked to each other in order to produce clear and readable code.

```go

func main(){
    model, err := NewSQLModel("MyTable", []string{"root:root@tcp(127.0.0.1:3306)/mydb", new(MySQLCnxOpenner))

    model.
    Select("a").
    SelectAvg("d").
    Where("a >=", "2").
    FindAll();
}

```

Structs are annoted with a `db:""` tag that make the mapping between the database schema and your go struct.

```go
type MyStuct struct {
    DbId             int     `db:id`
	ExportedInted    int     `db:"aaa"`
	ExportedString   string  `db:"bbb"`
	ExportedFloat64  float64 `db:"ccc"`
}
```

## Insert

```go
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

     model, err := NewSQLModel("MyTable", []string{"root:root@tcp(127.0.0.1:3306)/mydb", new(MySQLCnxOpenner))

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

     model, err := NewSQLModel("MyTable", []string{"root:root@tcp(127.0.0.1:3306)/mydb", new(MySQLCnxOpenner))

    myStruct := new(MyStruct)
    myStruct.DbId = 1;

    result := model.delete(myStruct)
    //Produces Delete from MyTable where id = 1
    fmt.Println(result)
    //Prints nil
    //The pointer is dereferenced when deleted
}
```


# Why ?

Why would you want an SQL wrapper in Go ? Well, don't you like to have enforced types and the additional safety that come with it ? Yes, then, this 
 
```go
stmtOut, err := db.Prepare("Select * from myTable")

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

- [x] documentation
- [ ] set of examples
- [x] Mysql
- [ ] Oracle
- [ ] PostgresSQL