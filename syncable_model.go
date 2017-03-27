package main

type SyncableModel interface {
	Find(id int)
	FindAll()
	FindUnionAll()
	FindAllBy(fields map[string]string, sqlType string)
	FindBy(field string, value string, sqlType string)
	Insert(fields map[string]string) (int, error)
	InsertBatch(fields []map[string]string)
	Update(where map[string]string, data map[string]string) (bool, error)
	Increment(where map[string]string, field string) (bool, error)
	Decrement(where map[string]string, field string) (bool, error)
	Delete(id int) (bool, error)
	DeleteWhere(where map[string]string) (bool, error)
	IsUnique(field string, value string) (bool, error)
	CountAll() (int, error)
	CountBy(field string, value string) (int, error)

	Where(field string, value string) (SyncableModel, error)
	OrderBy(field string, order string)

	SoftDelete(mode bool) SyncableModel
	AsArray() SyncableModel
	AsObject() SyncableModel
	AsJson() SyncableModel
	CreatedOn(row string) (string, error)
	ModifiedOn(row string) (string, error)
	RawSql(sql string)

	Select(selectString string) SyncableModel
	SelectMax(selectString string) SyncableModel
	SelectMin(selectString string) SyncableModel
	SelectAvg(selectString string) SyncableModel
	SelectSum(selectString string) SyncableModel
	Distinct(selectString string) SyncableModel

	From(from string) SyncableModel
	Join(table string, condition string, joinType string) SyncableModel
	Union(selectString string) SyncableModel
	OrWhere(key string, value string) SyncableModel
	WhereIn(key string, values []string) SyncableModel
	OrWhereIn(key string, values []string) SyncableModel
	WhereNotIn(key string, values []string) SyncableModel
	OrWhereNotIn(key string, values []string) SyncableModel
	Like(key string, value string) SyncableModel
	NotLike(key string, value string) SyncableModel
	OrLike(key string, value string) SyncableModel
	OrNotLike(key string, value string) SyncableModel
	GroupBy(key string) SyncableModel
	Having(key string, value string) SyncableModel
	OrHaving(key string, value string) SyncableModel
	Limit(number int) SyncableModel
	Offset(number int) SyncableModel
	Set(key string, value string) SyncableModel
	AffectedRows() (int, error)
	LastQuery() (bool, error)
	Truncate() (bool, error)
	InsertedID() (int, error)

	//getter
	Table() string
	LastError() error
	Key() string
	CreatedField() string
	ModifiedField() string
	DeletedField() string
	Created() bool
	Modified() bool
	Deleted() bool
	BeforeInsert() []func()
	AfterInsert() []func()
	BeforeUpdate() []func()
	AfterUpdate() []func()
	BeforeFind() []func()
	AfterFind() []func()
	BeforeUnionAll() []func()
	AfterUnionAll() []func()
	BeforeDelete() []func()
	AfterDelete() []func()
	ReturnType() string
	ReturnInsertID() string

	//setter
	SetTable(string) SyncableModel
	SetLastError(error) SyncableModel
	SetKey(string) SyncableModel
	SetCreatedField(string) SyncableModel
	SetModifiedField(string) SyncableModel
	SetDeletedField(string) SyncableModel
	SetCreated(bool) SyncableModel
	SetModified(bool) SyncableModel
	SetDeleted(bool) SyncableModel
	SetBeforeInsert([]func()) SyncableModel
	SetAfterInsert([]func()) SyncableModel
	SetBeforeUpdate([]func()) SyncableModel
	SetAfterUpdate([]func()) SyncableModel
	SetBeforeFind([]func()) SyncableModel
	SetAfterFind([]func()) SyncableModel
	SetBeforeUnionAll([]func()) SyncableModel
	SetAfterUnionAll([]func()) SyncableModel
	SetBeforeDelete([]func()) SyncableModel
	SetAfterDelete([]func()) SyncableModel
	SetReturnType(string) SyncableModel
	SetReturnInsertID(string) SyncableModel
}
