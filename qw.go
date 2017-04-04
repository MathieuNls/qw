package main

import (
	// connector package

	_ "github.com/mathieunls/qw/connector"
	// query package

	"fmt"

	query "github.com/mathieunls/qw/query"
)

func main() {

	solr, _ := query.NewSolrQuery("http://127.0.0.1:8983/solr/techproducts", 10, 10)

	type Test struct {
		ID    string `db:"id"`
		Price int    `db:"price"`
	}

	val, _ := solr.
		ReturnType(new(Test)).
		Select("id, price").
		Find("6H500F0")

	fmt.Println(val)

	/*one := val.(map[string]interface{})

	for key, value := range one {
		fmt.Println("Key:", key, "Value:", value, reflect.TypeOf(value))
	}

	two := one["response"].(map[string]interface{})["docs"].([]interface{})[0]

	three := two.(map[string]interface{})

	type Test struct {
		ID    string `db:"id"`
		Price int    `db:"price"`
	}

	t := new(Test)

	reflected := reflect.New(reflect.TypeOf(t).Elem()).Elem()
	typeOfT := reflected.Type()

	//For each field in the model.result
	for i := 0; i < reflected.NumField(); i++ {
		dbKey, _ := typeOfT.Field(i).Tag.Lookup("db")
		typeOfKey := reflected.Field(i).Type()

		switch typeOfKey.String() {
		case "int":
			intValue, _ := strconv.ParseInt(fmt.Sprintf("%v", three[dbKey]), 10, 64)
			reflected.Field(i).SetInt(intValue)
			break
		case "string":
			reflected.Field(i).SetString(fmt.Sprintf("%v", three[dbKey]))
			break
		case "float64":
			floatValue, _ := strconv.ParseFloat(fmt.Sprintf("%v", three[dbKey]), 64)
			reflected.Field(i).SetFloat(floatValue)
			break
		case "float32":
			floatValue, _ := strconv.ParseFloat(fmt.Sprintf("%v", three[dbKey]), 32)
			reflected.Field(i).SetFloat(floatValue)
			break
		}
	}

	fmt.Println(reflected.Interface())
	fmt.Println(three, reflect.TypeOf(three))*/

}
