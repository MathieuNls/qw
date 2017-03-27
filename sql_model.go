package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

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
	returnType string

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
	result interface{}

	mapping map[string]string
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
	model.mapping = make(map[string]string)

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

func (model SQLModel) reflectResult(values []sql.RawBytes, columns []string) interface{} {

	colMap := make(map[string]int)

	for index := 0; index < len(columns); index++ {
		colMap[columns[index]] = index
	}

	s := reflect.ValueOf(model.result).Elem()

	for k, v := range model.mapping {

		fmt.Println("Reflecting " + k)

		typeOfKey := s.FieldByName(k).Type()
		value := values[colMap[v]]

		switch typeOfKey.String() {
		case "int":
			fmt.Println("Reflecting " + v + " into " + k + ": " + string(value))
			intValue, _ := strconv.ParseInt(string(value), 10, 64)
			s.FieldByName(k).SetInt(intValue)
			break
		case "string":
			fmt.Println("Reflecting " + v + " into " + k + ": " + string(value))
			s.FieldByName(k).SetString(string(value))
			break
		}

	}

	return model.result
}

func (model SQLModel) Find(id int) (interface{}, error) {

	stmtOut, err := model.db.Prepare(
		"SELECT " + "*" +
			" FROM " + model.tableName +
			" WHERE " + model.key + " = ? LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()
	rows, err := stmtOut.Query(id)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		result := model.reflectResult(values, columns)

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		for i, col := range values {

			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			fmt.Println(columns[i], ": ", value)
		}
		fmt.Println("-----------------------------------")

		return result, nil
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nil, nil

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

func (model SQLModel) Select(selectString string) SyncableModel
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
