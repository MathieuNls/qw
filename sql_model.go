package gosqlwrapper

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"strings"
)

//SQLModel represents a SQLModel struct that offers helper function to safely
//interact with the database in go
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
	 * as_map() methods
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

	//The last query executed
	lastQuery string

	//mysql limit clause
	limit int

	//mysql offset clause
	offset int

	/**
	* Slices used to construct query
	 */
	pendingWheres  []string
	pendingJoins   []string
	pendingUnions  []string
	pendingGroupBy []string
	pendingHaving  []string
	pendingOrderBy []string
}

//NewSQLModel returns a pointer to a new SQLModel with all default values setted
func NewSQLModel(table string, dbCons []string, cnxOpener CnxOpener) (*SQLModel, error) {
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

	model.db, err = cnxOpener.OpenCnx(dbCons)

	if err != nil {
		return nil, err
	}

	return model, nil

}

//Key allow to modify the default id as pk for the table
func (model *SQLModel) Key(key string) *SQLModel {
	model.key = key
	return model
}

//Cleans up everything and make the model ready
//for a new request.
//It receives and stores the last error if needs be
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
			case "float64":
				floatValue, _ := strconv.ParseFloat(string(value), 64)
				reflected.Field(i).SetFloat(floatValue)
				break
			case "float32":
				floatValue, _ := strconv.ParseFloat(string(value), 32)
				reflected.Field(i).SetFloat(floatValue)
				break
			}
		}
	}

	return reflected.Interface()
}

// composeSelectString merges all the select clauses together
func (model *SQLModel) composeSelectString() string {
	selectString := "SELECT "

	if len(model.pendingSelects) > 0 {
		selectString += strings.Join(model.pendingSelects, ", ")
	} else {
		selectString += " * "
	}

	selectString += " FROM " + model.tableName

	if len(model.pendingJoins) > 0 {
		selectString += strings.Join(model.pendingJoins, " ")
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

// executeSelectQuery queries the database
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

// Debug prints all the select clauses to the consol
func (model *SQLModel) Debug() {

	fmt.Println("selects:" + strings.Join(model.pendingSelects, ", "))
	fmt.Println("wheres:" + strings.Join(model.pendingWheres, " AND "))
	fmt.Println("joins:" + strings.Join(model.pendingJoins, " "))
	fmt.Println("unions:" + strings.Join(model.pendingUnions, " "))
	fmt.Println("group by:" + strings.Join(model.pendingUnions, ", "))
}

// Join add a join clause to the ongoing select
//model.Join("myOtherTable", "mytable.id = myOtherTable.id", "left")
//will produce left join myOtherTable on mytable.id = myOtherTable.id
func (model *SQLModel) Join(table string, condition string, joinType string) *SQLModel {
	model.pendingJoins = append(model.pendingJoins, joinType+" JOIN "+" "+table+" ON "+condition)
	return model
}

// Union adds a union clause
func (model *SQLModel) Union(selectString string) *SQLModel {
	model.pendingUnions = append(model.pendingUnions, selectString)
	return model
}

// OrWhere adds a OrWhere clause
func (model *SQLModel) OrWhere(field string, value string) *SQLModel {
	return model.Where(" OR "+field, value)
}

// WhereIn adds a WhereIn clause
func (model *SQLModel) WhereIn(field string, values []string) *SQLModel {

	return model.Where(field+" IN", strings.Join(model.pendingSelects, ", "))
}

// OrWhereIn adds a OrWhereIn clause
func (model *SQLModel) OrWhereIn(field string, values []string) *SQLModel {

	return model.Where(" OR "+field+" IN", strings.Join(model.pendingSelects, ", "))
}

// WhereNotIn adds a WhereNotIn clause
func (model *SQLModel) WhereNotIn(field string, values []string) *SQLModel {

	return model.Where(field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

// OrWhereNotIn adds a OrWhereNotIn clause
func (model *SQLModel) OrWhereNotIn(field string, values []string) *SQLModel {

	return model.Where(" OR "+field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

// Like adds a Like clause
func (model *SQLModel) Like(field string, value string) *SQLModel {

	return model.Where(field+" LIKE", value)
}

// NotLike adds a NotLike clause
func (model *SQLModel) NotLike(field string, value string) *SQLModel {

	return model.Where(field+" NOT LIKE", value)
}

// OrLike adds a OrLike clause
func (model *SQLModel) OrLike(field string, value string) *SQLModel {

	return model.Where(" OR "+field+" LIKE", value)
}

// OrNotLike adds a OrNotLike clause
func (model *SQLModel) OrNotLike(field string, value string) *SQLModel {

	return model.Where(" OR "+field+" NOT LIKE", value)
}

// GroupBy adds a GroupBy clause
func (model *SQLModel) GroupBy(fields string) *SQLModel {

	model.pendingGroupBy = append(model.pendingGroupBy, fields)
	return model
}

// OrderBy adds a OrderBy clause
func (model *SQLModel) OrderBy(fields string, order string) *SQLModel {

	model.pendingOrderBy = append(model.pendingOrderBy, fields+" "+order)
	return model
}

// Having adds a Having clause
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

// OrHaving adds a OrHaving clause
func (model *SQLModel) OrHaving(field string, cond string) *SQLModel {
	return model.Having(" OR "+field, cond)
}

// Limit adds a Limit clause
func (model *SQLModel) Limit(limit int) *SQLModel {
	model.limit = limit
	return model
}

// Offset adds a Offset clause
func (model *SQLModel) Offset(offset int) *SQLModel {
	model.offset = offset
	return model
}

// LastQuery returns the last sql query executed
func (model *SQLModel) LastQuery() string {
	return model.lastQuery
}

// Where adds a Where clause
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

// Select adds a field to the select
func (model *SQLModel) Select(selectString string) *SQLModel {

	model.pendingSelects = append(model.pendingSelects, selectString)
	return model
}

// SelectMax adds a SelectMax field to the select
func (model *SQLModel) SelectMax(selectString string) *SQLModel {

	return model.Select("MAX(" + selectString + ")")
}

// SelectMin adds a SelectMin field to the select
func (model *SQLModel) SelectMin(selectString string) *SQLModel {

	return model.Select("MIN(" + selectString + ")")
}

// SelectAvg adds a SelectAvg field to the select
func (model *SQLModel) SelectAvg(selectString string) *SQLModel {

	return model.Select("AVG(" + selectString + ")")
}

// SelectSum adds a SelectSum field to the select
func (model *SQLModel) SelectSum(selectString string) *SQLModel {

	return model.Select("Sum(" + selectString + ")")
}

// Find returns the first row with key=id in a struct of ReturnType type
func (model *SQLModel) Find(id string) (interface{}, error) {

	err := model.
		Limit(1).
		Where(model.key, id).
		executeSelectQuery()

	return model.result[0], err
}

// FindAll returns all the row matching the query in an array of ReturnType type
func (model *SQLModel) FindAll() ([]interface{}, error) {

	err := model.
		executeSelectQuery()

	return model.result, err
}

// FindAllBy returns all the row matching the fields in an array of ReturnType type
func (model *SQLModel) FindAllBy(fields map[string]string) ([]interface{}, error) {

	for k, v := range fields {
		model.Where(k, v)
	}

	err := model.
		executeSelectQuery()

	return model.result, err
}

// FindBy returns all the first row matching the on ongoing select in a ReturnType struct
func (model *SQLModel) FindBy(field string, value string) (interface{}, error) {
	err := model.Where(field, value).executeSelectQuery()
	return model.result[0], err
}

// CountAll returns the number of rows in the table
func (model *SQLModel) CountAll() (int, error) {

	e := ""
	model.pendingSelects = []string{}
	err := model.
		Select(" count(1) ").
		db.QueryRow(model.composeSelectString()).Scan(&e)

	returnValue, _ := strconv.Atoi(e)
	return returnValue, err
}

// CountBy returns the number of rows in the table with field = value
func (model *SQLModel) CountBy(field string, value string) (int, error) {

	return model.
		Where(field, value).
		CountAll()
}

// IsUnique returns if field=value is unique in the db
func (model *SQLModel) IsUnique(field string, value string) (bool, error) {
	count, err := model.
		Where(field, value).
		CountAll()

	if count == 0 {
		return true, err
	}

	return false, err
}

// Insert insert a struct to the db
func (model *SQLModel) Insert(data interface{}) (bool, error) {

	columnString := []string{}
	var valueString []interface{}
	placeHolders := []string{}
	structPKIndex := -1

	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {

		column, dbTagPresent := typeOfT.Field(i).Tag.Lookup("db")

		if dbTagPresent && column != model.key {
			columnString = append(columnString, column)
			valueString = append(valueString, s.Field(i).Interface())
			placeHolders = append(placeHolders, "?")
		} else if column == model.key {
			structPKIndex = i
		}
	}

	insertStr := "INSERT INTO " + model.tableName +
		" (" + strings.Join(columnString, ", ") + ") " +
		" VALUES (" + strings.Join(placeHolders, ", ") + ")"

	stmtIns, err := model.db.Prepare(insertStr)

	model.lastQuery = insertStr

	if err != nil {
		return false, err
	}

	result, err := stmtIns.Exec(valueString...)
	lastInsertedID, err := result.LastInsertId()

	if err != nil {
		return false, err
	}

	if structPKIndex != -1 {
		s.Field(structPKIndex).SetInt(lastInsertedID)
	}

	return true, nil
}

// Delete deletes a struct from the db based on key
func (model *SQLModel) Delete(data interface{}) (bool, error) {

	structPKIndex := -1

	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {

		column, _ := typeOfT.Field(i).Tag.Lookup("db")

		if column == model.key {
			structPKIndex = i
			break
		}
	}

	pk := s.Field(structPKIndex).Interface().(int)

	deleteStr := "DELETE FROM " + model.tableName +
		" WHERE " + model.key + " = ?"

	stmtIns, err := model.db.Prepare(deleteStr)

	model.lastQuery = deleteStr

	if err != nil {
		return false, err
	}

	result, err := stmtIns.Exec(pk)
	lastInsertedID, err := result.LastInsertId()

	if err != nil {
		return false, err
	}

	if structPKIndex != -1 {
		s.Field(structPKIndex).SetInt(lastInsertedID)
	}

	data = nil

	return true, nil

}
