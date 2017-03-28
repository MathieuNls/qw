package main

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

	// fmt.Println(reflected)

	s := []string{
		"root:root@tcp(127.0.0.1:3306)/taxo",
	}
	model, err := NewSQLModel("bugs", s)

	model.returnType = new(A)

	model.Select("a, b, c").
	Select("d").
	Where("a", "b").


	// v, err := model.Select("INTERNAL_ID").
	// 	Select("EXTERNAL_ID").
	// 	CountAll()

	// fmt.Println(model.LastQuery())
	// fmt.Println(err)
	// fmt.Println(v)

	// for index := 0; index < len(v); index++ {
	// 	fmt.Println(v[index].(Bug).ExtID)
	// 	fmt.Println(reflect.TypeOf(v[index]))

	// }

}
