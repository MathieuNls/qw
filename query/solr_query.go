package query

import (
	"fmt"
	"reflect"
	"strconv"

	enginesolr "github.com/mathieunls/golr/engine"
	jsonsolr "github.com/mathieunls/golr/solrjson"
)

type SolrQuery struct {
	query

	engine      *enginesolr.Solr
	jsonBuilder *jsonsolr.SolrJSONBuilder

	pendingStatsFields []string
	pendingStatsMax    []string
	pendingStatsAVG    []string
	pendingStatsSum    []string
	pendingStatsMin    []string
}

func NewSolrQuery(url string, timeout int, commit int) (*SolrQuery, error) {

	solrQuery := new(SolrQuery)

	solrQuery.engine = enginesolr.NewSolr(url, timeout, commit)
	solrQuery.jsonBuilder = jsonsolr.NewSolrJSONBuilder()
	solrQuery.pendingWheres = []string{}
	solrQuery.pendingSelects = []string{}
	solrQuery.pendingStatsFields = []string{}
	solrQuery.pendingStatsMax = []string{}
	solrQuery.pendingStatsAVG = []string{}
	solrQuery.pendingStatsSum = []string{}
	solrQuery.pendingStatsMin = []string{}
	solrQuery.key = "id"

	return solrQuery, nil
}

func (solrQuery *SolrQuery) Debug() {

	fmt.Println(string(solrQuery.jsonBuilder.Prepare()))
}

func (solrQuery *SolrQuery) Join(table string, condition string, joinType string) *SolrQuery {
	panic("Unsupported")
}

func (solrQuery *SolrQuery) Union(selectString string) *SolrQuery {
	panic("Unsupported")
}

func (solrQuery *SolrQuery) OrWhere(field string, value string) *SolrQuery {

	solrQuery.jsonBuilder.Filter(field, value)
	return solrQuery
}

func (solrQuery *SolrQuery) WhereIn(field string, values []string) *SolrQuery {

	for index := 0; index < len(values); index++ {
		solrQuery.jsonBuilder.Filter(field, values[index])
	}
	return solrQuery
}

func (solrQuery *SolrQuery) OrWhereIn(field string, values []string) *SolrQuery {
	for index := 0; index < len(values); index++ {
		solrQuery.jsonBuilder.Filter(" OR "+field, values[index])
	}
	return solrQuery
}

func (solrQuery *SolrQuery) WhereNotIn(field string, values []string) *SolrQuery {
	for index := 0; index < len(values); index++ {
		solrQuery.jsonBuilder.Filter(" -"+field, values[index])
	}
	return solrQuery
}

func (solrQuery *SolrQuery) OrWhereNotIn(field string, values []string) *SolrQuery {
	for index := 0; index < len(values); index++ {
		solrQuery.jsonBuilder.Filter(" OR -"+field, values[index])
	}
	return solrQuery
}

func (solrQuery *SolrQuery) Like(field string, value string) *SolrQuery {
	solrQuery.jsonBuilder.Filter(field, value+"~2")
	return solrQuery
}

func (solrQuery *SolrQuery) NotLike(field string, value string) *SolrQuery {
	solrQuery.jsonBuilder.Filter(" -"+field, value+"~2")
	return solrQuery
}

func (solrQuery *SolrQuery) OrLike(field string, value string) *SolrQuery {
	solrQuery.jsonBuilder.Filter(" OR "+field, value+"~2")
	return solrQuery
}

func (solrQuery *SolrQuery) OrNotLike(field string, value string) *SolrQuery {
	solrQuery.jsonBuilder.Filter(" OR -"+field, value+"~2")
	return solrQuery
}

func (solrQuery *SolrQuery) GroupBy(fields string) *SolrQuery {
	panic("Unsupported")
}

func (solrQuery *SolrQuery) OrderBy(fields string, order string) *SolrQuery {
	solrQuery.jsonBuilder.Sort(fields, order)
	return solrQuery
}

func (solrQuery *SolrQuery) Having(field string, cond string) *SolrQuery {
	panic("Unsupported")
}

func (solrQuery *SolrQuery) OrHaving(field string, cond string) *SolrQuery {
	panic("Unsupported")
}

func (solrQuery *SolrQuery) Limit(limit int) *SolrQuery {
	solrQuery.jsonBuilder.Limit(limit)
	return solrQuery
}

func (solrQuery *SolrQuery) Offset(offset int) *SolrQuery {
	solrQuery.jsonBuilder.Offset(offset)
	return solrQuery
}

func (solrQuery *SolrQuery) Where(field string, value string) *SolrQuery {
	solrQuery.pendingWheres = append(solrQuery.pendingWheres, " "+field+":\""+value+"\"")
	return solrQuery
}

func (solrQuery *SolrQuery) Select(selectString string) *SolrQuery {
	solrQuery.jsonBuilder.Field(selectString)
	return solrQuery
}

func (solrQuery *SolrQuery) SelectMax(selectString string) *SolrQuery {
	solrQuery.pendingStatsFields = append(solrQuery.pendingStatsFields, selectString)
	solrQuery.pendingStatsMax = append(solrQuery.pendingStatsMax, selectString)
	return solrQuery
}

func (solrQuery *SolrQuery) SelectMin(selectString string) *SolrQuery {
	solrQuery.pendingStatsFields = append(solrQuery.pendingStatsFields, selectString)
	solrQuery.pendingStatsMin = append(solrQuery.pendingStatsMin, selectString)
	return solrQuery
}

func (solrQuery *SolrQuery) SelectAvg(selectString string) *SolrQuery {
	solrQuery.pendingStatsFields = append(solrQuery.pendingStatsFields, selectString)
	solrQuery.pendingStatsAVG = append(solrQuery.pendingStatsAVG, selectString)
	return solrQuery
}

func (solrQuery *SolrQuery) SelectSum(selectString string) *SolrQuery {
	solrQuery.pendingStatsFields = append(solrQuery.pendingStatsFields, selectString)
	solrQuery.pendingStatsSum = append(solrQuery.pendingStatsSum, selectString)
	return solrQuery
}

func (solrQuery *SolrQuery) Find(id string) (interface{}, error) {

	solrQuery.jsonBuilder.Filter(solrQuery.key, id)

	val, err := solrQuery.engine.Query(solrQuery.jsonBuilder)

	toMap := val.(map[string]interface{})
	docs := toMap["response"].(map[string]interface{})["docs"].([]interface{})

	for index := 0; index < len(docs); index++ {

		solrQuery.result = append(solrQuery.result, solrQuery.reflectValues(docs[index].(map[string]interface{})))
	}

	return solrQuery.result, err
}

func (solrQuery *SolrQuery) reflectValues(values map[string]interface{}) interface{} {

	reflected := reflect.New(reflect.TypeOf(solrQuery.returnType).Elem()).Elem()
	typeOfT := reflected.Type()

	//For each field in the model.result
	for i := 0; i < reflected.NumField(); i++ {
		dbKey, _ := typeOfT.Field(i).Tag.Lookup("db")
		typeOfKey := reflected.Field(i).Type()

		switch typeOfKey.String() {
		case "int":
			intValue, _ := strconv.ParseInt(fmt.Sprintf("%v", values[dbKey]), 10, 64)
			reflected.Field(i).SetInt(intValue)
			break
		case "string":
			reflected.Field(i).SetString(fmt.Sprintf("%v", values[dbKey]))
			break
		case "float64":
			floatValue, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[dbKey]), 64)
			reflected.Field(i).SetFloat(floatValue)
			break
		case "float32":
			floatValue, _ := strconv.ParseFloat(fmt.Sprintf("%v", values[dbKey]), 32)
			reflected.Field(i).SetFloat(floatValue)
			break
		}
	}

	return reflected.Interface()

}

func (model *SolrQuery) ReturnType(returnType interface{}) *SolrQuery {
	model.returnType = returnType
	return model
}

// func (solrQuery *SolrQuery) FindAll() ([]interface{}, error) {

// }
// func (solrQuery *SolrQuery) FindAllBy(fields map[string]string) ([]interface{}, error) {
// }
// func (solrQuery *SolrQuery) FindBy(field string, value string) (interface{}, error) {
// }
// func (solrQuery *SolrQuery) CountAll() (int, error) {
// }
// func (solrQuery *SolrQuery) CountBy(field string, value string) (int, error) {
// }
// func (solrQuery *SolrQuery) IsUnique(field string, value string) (bool, error) {
// }
// func (solrQuery *SolrQuery) Insert(data interface{}) (bool, error) {
// }
// func (solrQuery *SolrQuery) Delete(data interface{}) (bool, error) {
// }
// func (solrQuery *SolrQuery) Update(data interface{}) (bool, error) {

// }
