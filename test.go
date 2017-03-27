package main

import "fmt"

type E struct {
	ID    int
	ExtID string
}

type T struct {
	A int
	B string
}

func main() {
	s := make([]string, 1)
	s[0] = "root:root@tcp(192.168.0.112:3306)/taxo"
	model, err := NewSQLModel("bugs", s)

	if err != nil {
		panic(err.Error())
	}
	model.key = "INTERNAL_ID"

	model.mapping["ID"] = "INTERNAL_ID"
	model.mapping["ExtID"] = "EXTERNAL_ID"
	model.result = new(E)

	model.Find(1)

	fmt.Println(model.result.(*E).ExtID)

	e := model.result.(*E)
	fmt.Println(e.ExtID)

}
