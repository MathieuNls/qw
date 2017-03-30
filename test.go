package main

import "fmt"

type Bug struct {
	ID    int    `db:"INTERNAL_ID"`
	ExtID string `db:"EXTERNAL_ID"`
}

type A struct {
	A int `db:"INTERNAL_ID"`
	B int `db:"EXTERNAL_ID"`
}

func main() {

	// real := new(A)
	// reflected := reflect.New(reflect.TypeOf(real).Elem()).Elem()
	// fmt.Println(real)
	// fmt.Println(reflected)
	// typeOfT := reflected.Type()

	// for i := 0; i < reflected.NumField(); i++ {
	// 	fmt.Println(i)
	// 	fmt.Println(typeOfT.Field(i))
	// 	fmt.Println(typeOfT.Field(i).Tag)

	// 	reflected.Field(i).SetInt(int64(i))
	// }

	// // fmt.Println(reflected)

	s := []string{
		"root:root@tcp(127.0.0.1:3306)/taxo",
	}
	model, err := NewSQLModel("bugs", s, new(MySQLCnxOpenner))

	if err != nil {
		panic("1")
	}

	model.returnType = new(Bug)

	v, err := model.Select("INTERNAL_ID").
		Select("EXTERNAL_ID").
		FindAll()

	fmt.Println(model.LastQuery())

	// fmt.Println(model.LastQuery())
	fmt.Println(v)
	fmt.Println(err)
	fmt.Println(v[0].(Bug))

	// for index := 0; index < len(v); index++ {
	// 	fmt.Println(v[index].(Bug).ExtID)
	// 	fmt.Println(reflect.TypeOf(v[index]))

	// }

}
