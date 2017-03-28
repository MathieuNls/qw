package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type SQLModel struct {

	//The name of the db table this model primarily uses.
	tableName string

	// Stores custom errors that can be reported.
	lastError error

	//The primary key of the table. Used as the 'id' throughout.
	key string

	// Field name to use for the created time column in the DB table if
	// setCreated is enabled
	createdField string

	// Field name to use for the modified time column in the DB table if
	// setModified is enabled
	modifiedField string

	// Field name to use for the deleted column in the DB table if softDeletes
	// is enabled
	deletedField string

	// Whether or not to auto-fill the createdField on inserts.
	setCreated bool

	// Whether or not to auto-fill the modifiedField on updates.
	setModified bool

	// The type of date/time field used for createdField and modifiedField
	// Valid values are 'int', 'datetime', 'date'
	dateFormat string

	// If false, the delete() method will perform a delete of that row.
	// If true, the value in deletedField will be set to 1.
	softDeletes bool

	// Stores any selects here for use by the find* functions.
	selects string

	// If false, the select() method will not try to protect your field or table
	// names with backticks.
	// This is useful if you need a compound select statement.
	escape bool

	// DB Connection details with fallback connections
	// each connections will be tried sequentially
	dbCon []string

	/**
	* Observer Slices
	*
	* Each slice can contain callback functions
	* which will be called during each event.
	*
	* <code>
	*  model.BeforeInsert(
	*    []func{
	*        func() string{ return "Hello" },
	*        func() string{ return " " },
	*        func() string{ return "World" }
	*   }
	* )
	* </code>
	**/

	beforeInsert   []func()
	afterInsert    []func()
	beforeUpdate   []func()
	afterUpdate    []func()
	beforeFind     []func()
	afterFind      []func()
	beforeUnionAll []func()
	afterUnionAll  []func()
	beforeDelete   []func()
	afterDelete    []func()

	/**
	 * By default, we return items as objects. You can change this for the
	 * entire class by setting this value to 'array' instead of 'object'.
	 * Alternatively, you can do it on a per-instance basis using the
	 * 'as_array()' and 'as_object()' methods.
	 */
	returnType interface{}

	/**
	 * Holds the return type temporarily when using the
	 * as_array() and as_object() methods
	 */
	tmpReturnType string

	/**
	 * If true, inserts will return the inserted ID.
	 *
	 * This can potentially slow down large imports drastically, so you can turn
	 * it off with the ReturnInsertID(false) method.
	 */
	returnInsertID bool

	// This array will be populated with selects and mainly use for union based query.
	pendingSelects []string

	//actual database connection
	db *sql.DB

	//return struct holder
	result []interface{}

	lastQuery string

	limit int

	offset int

	pendingWheres []string

	pendingJoins []string

	pendingUnions []string

	pendingGroupBy []string

	pendingHaving []string

	pendingOrderBy []string
}

func NewSQLModel(table string, dbCons []string) (*SQLModel, error) {
	model := new(SQLModel)
	model.tableName = table
	model.key = "id"
	model.createdField = "created_on"
	model.modifiedField = "modified_on"
	model.deletedField = "deleted"
	model.setCreated = false
	model.setModified = false
	model.softDeletes = false
	model.dateFormat = "datetime"
	model.pendingSelects = []string{}
	model.pendingWheres = []string{}
	model.pendingJoins = []string{}
	model.pendingUnions = []string{}
	model.pendingGroupBy = []string{}
	model.pendingOrderBy = []string{}
	model.lastQuery = ""
	model.limit = -1
	model.offset = -1

	var err error

	for index := 0; index < len(dbCons); index++ {
		//Check dns format
		model.db, err = sql.Open("mysql", dbCons[index])
		if err != nil {
			fmt.Println(dbCons[index] + " failed to open")
		} else {
			//Check database connectivity
			err = model.db.Ping()
			if err != nil {
				fmt.Println(dbCons[index] + " failed to answer ping")
			} else {
				//Database is answering, break here
				break
			}
			defer model.db.Close()

		}
	}

	if err != nil {
		return nil, err
	}

	return model, nil

}

func (model *SQLModel) cleanup(err error) {
	model.pendingSelects = []string{}
	model.pendingWheres = []string{}
	model.pendingJoins = []string{}
	model.pendingUnions = []string{}
	model.pendingGroupBy = []string{}
	model.pendingOrderBy = []string{}
	model.limit = -1
	model.offset = -1
	model.lastError = err
}

// Adapt []sql.RawBytes to the model.result struct using `db` Tag
func (model *SQLModel) reflectResult(values []sql.RawBytes, columns []string) interface{} {

	colMap := make(map[string]int)

	//Index the columns by name for O(1) access
	for index := 0; index < len(columns); index++ {
		colMap[columns[index]] = index
	}

	//Reflect on model.returnType for reading/writing on fields
	reflected := reflect.New(reflect.TypeOf(model.returnType).Elem()).Elem()
	typeOfT := reflected.Type()

	//For each field in the model.result
	for i := 0; i < reflected.NumField(); i++ {

		dbKey, _ := typeOfT.Field(i).Tag.Lookup("db")

		//Check if that tag is present in the resultset
		if valKey, ok := colMap[dbKey]; ok {

			//Get the value from the resultset
			value := values[valKey]

			//Swith on target type for byte to type convertion
			typeOfKey := reflected.Field(i).Type()
			switch typeOfKey.String() {
			case "int":
				intValue, _ := strconv.ParseInt(string(value), 10, 64)
				reflected.Field(i).SetInt(intValue)
				break
			case "string":
				reflected.Field(i).SetString(string(value))
				break
			case "float":
				floatValue, _ := strconv.ParseFloat(string(value), 32)
				reflected.Field(i).SetFloat(floatValue)
				break
			}
		}
	}

	return reflected.Interface()
}

func (model *SQLModel) composeSelectString() string {
	selectString := "SELECT "

	if len(model.pendingSelects) > 0 {
		selectString += strings.Join(model.pendingSelects, ", ")
	} else {
		selectString += " * "
	}

	selectString += " FROM " + model.tableName

	if len(model.pendingJoins) > 0 {
		selectString += " JOIN " + strings.Join(model.pendingJoins, ", ")
	}

	if len(model.pendingWheres) > 0 {
		selectString += " WHERE " + strings.Join(model.pendingWheres, " ")
	}

	if len(model.pendingGroupBy) > 0 {
		selectString += " GROUP BY " + strings.Join(model.pendingGroupBy, ", ")
	}

	if len(model.pendingHaving) > 0 {
		selectString += " HAVING " + strings.Join(model.pendingHaving, " ")
	}

	if len(model.pendingGroupBy) > 0 {
		selectString += " ORDER BY " + strings.Join(model.pendingGroupBy, ", ")
	}

	selectString += strings.Join(model.pendingUnions, " ")

	if model.limit > 0 {
		selectString += " LIMIT " + strconv.Itoa(model.limit)
	}

	if model.offset > 0 {
		selectString += " OFFSET  " + strconv.Itoa(model.offset)
	}

	model.lastQuery = selectString
	return selectString
}

func (model *SQLModel) executeSelectQuery() error {

	selectString := model.composeSelectString()
	stmtOut, err := model.db.Prepare(selectString)
	model.lastQuery = selectString

	if err != nil {
		model.cleanup(err)
		return err
	}
	defer stmtOut.Close()
	rows, err := stmtOut.Query()

	columns, err := rows.Columns()
	if err != nil {
		model.cleanup(err)
		return err
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
			model.cleanup(err)
			return err
		}

		model.result = append(model.result, model.reflectResult(values, columns))
	}
	if err = rows.Err(); err != nil {
		model.cleanup(err)
		return err
	}
	model.cleanup(nil)
	return nil
}

func (model *SQLModel) Debug() {

	fmt.Println("selects:" + strings.Join(model.pendingSelects, ", "))
	fmt.Println("wheres:" + strings.Join(model.pendingWheres, " AND "))
	fmt.Println("joins:" + strings.Join(model.pendingJoins, " "))
	fmt.Println("unions:" + strings.Join(model.pendingUnions, " "))
	fmt.Println("group by:" + strings.Join(model.pendingUnions, ", "))
}

func (model *SQLModel) Join(table string, condition string, joinType string) *SQLModel {
	model.pendingJoins = append(model.pendingJoins, joinType+" "+table+" ON "+condition)
	return model
}

func (model *SQLModel) Union(selectString string) *SQLModel {
	model.pendingUnions = append(model.pendingUnions, selectString)
	return model
}

func (model *SQLModel) OrWhere(field string, value string) *SQLModel {
	return model.Where(" OR "+field, value)
}

func (model *SQLModel) WhereIn(field string, values []string) *SQLModel {

	return model.Where(field+" IN", strings.Join(model.pendingSelects, ", "))
}

func (model *SQLModel) OrWhereIn(field string, values []string) *SQLModel {

	return model.Where(" OR "+field+" IN", strings.Join(model.pendingSelects, ", "))
}

func (model *SQLModel) WhereNotIn(field string, values []string) *SQLModel {

	return model.Where(field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

func (model *SQLModel) OrWhereNotIn(field string, values []string) *SQLModel {

	return model.Where(" OR "+field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

func (model *SQLModel) Like(field string, value string) *SQLModel {

	return model.Where(field+" LIKE", value)
}

func (model *SQLModel) NotLike(field string, value string) *SQLModel {

	return model.Where(field+" NOT LIKE", value)
}

func (model *SQLModel) OrLike(field string, value string) *SQLModel {

	return model.Where(" OR "+field+" LIKE", value)
}

func (model *SQLModel) OrNotLike(field string, value string) *SQLModel {

	return model.Where(" OR "+field+" NOT LIKE", value)
}

func (model *SQLModel) GroupBy(fields string) *SQLModel {

	model.pendingGroupBy = append(model.pendingGroupBy, fields)
	return model
}

func (model *SQLModel) OrderBy(fields string, order string) *SQLModel {

	model.pendingOrderBy = append(model.pendingOrderBy, fields+" "+order)
	return model
}

func (model *SQLModel) Having(field string, cond string) *SQLModel {

	concatAnd := func(field string) string {
		if len(model.pendingHaving) > 0 && !strings.HasPrefix(field, " OR ") {
			return " AND " + field
		}
		return field
	}

	model.pendingHaving = append(model.pendingHaving, concatAnd(field)+" "+cond)
	return model
}

func (model *SQLModel) OrHaving(field string, cond string) *SQLModel {
	return model.Having(" OR "+field, cond)
}

func (model *SQLModel) Limit(limit int) *SQLModel {
	model.limit = limit
	return model
}

func (model *SQLModel) Offset(offset int) *SQLModel {
	model.offset = offset
	return model
}

func (model *SQLModel) LastQuery() string {
	return model.lastQuery
}

func (model *SQLModel) Where(field string, value string) *SQLModel {

	concatAnd := func(field string) string {
		if len(model.pendingWheres) > 0 && !strings.HasPrefix(field, " OR ") {
			return " AND " + field
		}
		return field
	}

	stringifyValue := func(value string) string {
		if _, err := strconv.Atoi(value); err == nil {
			return value
		} else if _, err := strconv.ParseFloat(value, 64); err == nil {
			return value
		} else if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
			return value
		}
		return "'" + value + "'"
	}

	hasSpecialSuffix := false
	specialSuffixes := []string{">=", ">", " <=", " <", " !=", " <>", " NOT LIKE", " LIKE", " NOT IN", " IN"}

	for i := 0; i < len(specialSuffixes); i++ {
		if strings.HasSuffix(field, specialSuffixes[i]) {

			hasSpecialSuffix = true
			break
		}
	}

	if hasSpecialSuffix {
		model.pendingWheres = append(model.pendingWheres, concatAnd(field)+" "+stringifyValue(value))
	} else {
		model.pendingWheres = append(model.pendingWheres, concatAnd(field)+" = "+stringifyValue(value))
	}

	return model
}

func (model *SQLModel) Select(selectString string) *SQLModel {

	model.pendingSelects = append(model.pendingSelects, selectString)
	return model
}

func (model *SQLModel) Find(id string) (interface{}, error) {

	err := model.
		Limit(1).
		Where(model.key, id).
		executeSelectQuery()

	return model.result[0], err
}

func (model SQLModel) FindAll() ([]interface{}, error) {

	err := model.
		executeSelectQuery()

	return model.result, err
}

/*
func (model SQLModel) FindAll()
func (model SQLModel) FindUnionAll()
func (model SQLModel) FindAllBy(fields map[string]string, sqlType string)
func (model SQLModel) FindBy(field string, value string, sqlType string)
func (model SQLModel) Insert(fields map[string]string) (int, error)
func (model SQLModel) InsertBatch(fields []map[string]string)
func (model SQLModel) Update(where map[string]string, data map[string]string) (bool, error)
func (model SQLModel) Increment(where map[string]string, field string) (bool, error)
func (model SQLModel) Decrement(where map[string]string, field string) (bool, error)
func (model SQLModel) Delete(id int) (bool, error)
func (model SQLModel) DeleteWhere(where map[string]string) (bool, error)
func (model SQLModel) IsUnique(field string, value string) (bool, error)
func (model SQLModel) CountAll() (int, error)
func (model SQLModel) CountBy(field string, value string) (int, error)

func (model SQLModel) Where(field string, value string) (SyncableModel, error)
func (model SQLModel) OrderBy(field string, order string)

func (model SQLModel) SoftDelete(mode bool) SyncableModel
func (model SQLModel) AsArray() SyncableModel
func (model SQLModel) AsObject() SyncableModel
func (model SQLModel) AsJson() SyncableModel
func (model SQLModel) CreatedOn(row string) (string, error)
func (model SQLModel) ModifiedOn(row string) (string, error)
func (model SQLModel) RawSql(sql string)


func (model SQLModel) SelectMax(selectString string) SyncableModel
func (model SQLModel) SelectMin(selectString string) SyncableModel
func (model SQLModel) SelectAvg(selectString string) SyncableModel
func (model SQLModel) SelectSum(selectString string) SyncableModel
func (model SQLModel) Distinct(selectString string) SyncableModel

func (model SQLModel) From(from string) SyncableModel
func (model SQLModel) Join(table string, condition string, joinType string) SyncableModel
func (model SQLModel) Union(selectString string) SyncableModel
func (model SQLModel) OrWhere(key string, value string) SyncableModel
func (model SQLModel) WhereIn(key string, values []string) SyncableModel
func (model SQLModel) OrWhereIn(key string, values []string) SyncableModel
func (model SQLModel) WhereNotIn(key string, values []string) SyncableModel
func (model SQLModel) OrWhereNotIn(key string, values []string) SyncableModel
func (model SQLModel) Like(key string, value string) SyncableModel
func (model SQLModel) NotLike(key string, value string) SyncableModel
func (model SQLModel) OrLike(key string, value string) SyncableModel
func (model SQLModel) OrNotLike(key string, value string) SyncableModel
func (model SQLModel) GroupBy(key string) SyncableModel
func (model SQLModel) Having(key string, value string) SyncableModel
func (model SQLModel) OrHaving(key string, value string) SyncableModel
func (model SQLModel) Limit(number int) SyncableModel
func (model SQLModel) Offset(number int) SyncableModel
func (model SQLModel) Set(key string, value string) SyncableModel
func (model SQLModel) AffectedRows() (int, error)
func (model SQLModel) LastQuery() (bool, error)
func (model SQLModel) Truncate() (bool, error)
func (model SQLModel) InsertedID() (int, error)

func (model SQLModel) Table() string
func (model SQLModel) LastError() error
func (model SQLModel) Key() string
func (model SQLModel) CreatedField() string
func (model SQLModel) ModifiedField() string
func (model SQLModel) DeletedField() string
func (model SQLModel) Created() bool
func (model SQLModel) Modified() bool
func (model SQLModel) Deleted() bool
func (model SQLModel) BeforeInsert() []func()
func (model SQLModel) AfterInsert() []func()
func (model SQLModel) BeforeUpdate() []func()
func (model SQLModel) AfterUpdate() []func()
func (model SQLModel) BeforeFind() []func()
func (model SQLModel) AfterFind() []func()
func (model SQLModel) BeforeUnionAll() []func()
func (model SQLModel) AfterUnionAll() []func()
func (model SQLModel) BeforeDelete() []func()
func (model SQLModel) AfterDelete() []func()
func (model SQLModel) ReturnType() string
func (model SQLModel) ReturnInsertID() string

func (model SQLModel) SetTable(string) SyncableModel
func (model SQLModel) SetLastError(error) SyncableModel
func (model SQLModel) SetKey(string) SyncableModel
func (model SQLModel) SetCreatedField(string) SyncableModel
func (model SQLModel) SetModifiedField(string) SyncableModel
func (model SQLModel) SetDeletedField(string) SyncableModel
func (model SQLModel) SetCreated(bool) SyncableModel
func (model SQLModel) SetModified(bool) SyncableModel
func (model SQLModel) SetDeleted(bool) SyncableModel
func (model SQLModel) SetBeforeInsert([]func()) SyncableModel
func (model SQLModel) SetAfterInsert([]func()) SyncableModel
func (model SQLModel) SetBeforeUpdate([]func()) SyncableModel
func (model SQLModel) SetAfterUpdate([]func()) SyncableModel
func (model SQLModel) SetBeforeFind([]func()) SyncableModel
func (model SQLModel) SetAfterFind([]func()) SyncableModel
func (model SQLModel) SetBeforeUnionAll([]func()) SyncableModel
func (model SQLModel) SetAfterUnionAll([]func()) SyncableModel
func (model SQLModel) SetBeforeDelete([]func()) SyncableModel
func (model SQLModel) SetAfterDelete([]func()) SyncableModel
func (model SQLModel) SetReturnType(string) SyncableModel
func (model SQLModel) SetReturnInsertID(string) SyncableModel
**/
