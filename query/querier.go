package query

//Querier represents whats doable accross all adapators
//Any new adaptor must implement this
type Querier interface {
	Key(key string) *Querier
	CreatedField(createdField string) *Querier
	ModifiedField(modifiedField string) *Querier
	DeletedField(deletedField string) *Querier
	Created(created bool) *Querier
	Modified(modified bool) *Querier
	SoftDeletes(softDeletes bool) *Querier
	DateFormat(dateFormat string) *Querier
	Debug()
	Join(table string, condition string, joinType string) *Querier
	Union(selectString string) *Querier
	OrWhere(field string, value string) *Querier
	WhereIn(field string, values []string) *Querier
	OrWhereIn(field string, values []string) *Querier
	WhereNotIn(field string, values []string) *Querier
	OrWhereNotIn(field string, values []string) *Querier
	Like(field string, value string) *Querier
	NotLike(field string, value string) *Querier
	OrLike(field string, value string) *Querier
	OrNotLike(field string, value string) *Querier
	GroupBy(fields string) *Querier
	OrderBy(fields string, order string) *Querier
	Having(field string, cond string) *Querier
	OrHaving(field string, cond string) *Querier
	Limit(limit int) *Querier
	Offset(offset int) *Querier
	LastQuery() string
	Where(field string, value string) *Querier
	Select(selectString string) *Querier
	SelectMax(selectString string) *Querier
	SelectMin(selectString string) *Querier
	SelectAvg(selectString string) *Querier
	SelectSum(selectString string) *Querier
	Find(id string) (interface{}, error)
	FindAll() ([]interface{}, error)
	FindAllBy(fields map[string]string) ([]interface{}, error)
	FindBy(field string, value string) (interface{}, error)
	CountAll() (int, error)
	CountBy(field string, value string) (int, error)
	IsUnique(field string, value string) (bool, error)
	Insert(data interface{}) (bool, error)
	Delete(data interface{}) (bool, error)
	Update(data interface{}) (bool, error)
	BeforeInsert(triggers []func([]interface{})) *Querier
	AfterInsert(triggers []func([]interface{})) *Querier
	BeforeUpdate(triggers []func([]interface{})) *Querier
	AfterUpdate(triggers []func([]interface{})) *Querier
	BeforeFind(triggers []func([]interface{})) *Querier
	AfterFind(triggers []func([]interface{})) *Querier
	BeforeDelete(triggers []func([]interface{})) *Querier
	AfterDelete(triggers []func([]interface{})) *Querier
}
