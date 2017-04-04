package query

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"

	"strings"

	"github.com/mathieunls/qw/connector"
)

//SQLQuery represents a SQLQuery struct that offers helper function to safely
//interact with the database in go
type SQLQuery struct {
	query

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

	beforeInsert []func([]interface{})
	afterInsert  []func([]interface{})
	beforeUpdate []func([]interface{})
	afterUpdate  []func([]interface{})
	beforeFind   []func([]interface{})
	afterFind    []func([]interface{})
	beforeDelete []func([]interface{})
	afterDelete  []func([]interface{})

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

//NewSQLQuery returns a pointer to a new SQLModel with all default values setted
func NewSQLQuery(table string, dbCons []string, cnxOpener connector.Cnx) (*SQLQuery, error) {
	model := new(SQLQuery)
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

//Cleans up everything and make the model ready
//for a new request.
//It receives and stores the last error if needs be
func (model *SQLQuery) cleanup(err error) {
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
func (model *SQLQuery) reflectResult(values []sql.RawBytes, columns []string) interface{} {

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
func (model *SQLQuery) composeSelectString() string {
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
func (model *SQLQuery) executeSelectQuery() error {

	model.executebeforeInsert()

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

	model.executeafterInsert()

	model.cleanup(nil)

	return nil
}

// Debug prints all the select clauses to the consol
func (model *SQLQuery) Debug() {

	fmt.Println("selects:" + strings.Join(model.pendingSelects, ", "))
	fmt.Println("wheres:" + strings.Join(model.pendingWheres, " AND "))
	fmt.Println("joins:" + strings.Join(model.pendingJoins, " "))
	fmt.Println("unions:" + strings.Join(model.pendingUnions, " "))
	fmt.Println("group by:" + strings.Join(model.pendingUnions, ", "))
}

// Join add a join clause to the ongoing select
//model.Join("myOtherTable", "mytable.id = myOtherTable.id", "left")
//will produce left join myOtherTable on mytable.id = myOtherTable.id
func (model *SQLQuery) Join(table string, condition string, joinType string) *SQLQuery {
	model.pendingJoins = append(model.pendingJoins, joinType+" JOIN "+" "+table+" ON "+condition)
	return model
}

// Union adds a union clause
func (model *SQLQuery) Union(selectString string) *SQLQuery {
	model.pendingUnions = append(model.pendingUnions, selectString)
	return model
}

// OrWhere adds a OrWhere clause
func (model *SQLQuery) OrWhere(field string, value string) *SQLQuery {
	return model.Where(" OR "+field, value)
}

// WhereIn adds a WhereIn clause
func (model *SQLQuery) WhereIn(field string, values []string) *SQLQuery {

	return model.Where(field+" IN", strings.Join(model.pendingSelects, ", "))
}

// OrWhereIn adds a OrWhereIn clause
func (model *SQLQuery) OrWhereIn(field string, values []string) *SQLQuery {

	return model.Where(" OR "+field+" IN", strings.Join(model.pendingSelects, ", "))
}

// WhereNotIn adds a WhereNotIn clause
func (model *SQLQuery) WhereNotIn(field string, values []string) *SQLQuery {

	return model.Where(field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

// OrWhereNotIn adds a OrWhereNotIn clause
func (model *SQLQuery) OrWhereNotIn(field string, values []string) *SQLQuery {

	return model.Where(" OR "+field+" NOT IN", strings.Join(model.pendingSelects, ", "))
}

// Like adds a Like clause
func (model *SQLQuery) Like(field string, value string) *SQLQuery {

	return model.Where(field+" LIKE", value)
}

// NotLike adds a NotLike clause
func (model *SQLQuery) NotLike(field string, value string) *SQLQuery {

	return model.Where(field+" NOT LIKE", value)
}

// OrLike adds a OrLike clause
func (model *SQLQuery) OrLike(field string, value string) *SQLQuery {

	return model.Where(" OR "+field+" LIKE", value)
}

// OrNotLike adds a OrNotLike clause
func (model *SQLQuery) OrNotLike(field string, value string) *SQLQuery {

	return model.Where(" OR "+field+" NOT LIKE", value)
}

// GroupBy adds a GroupBy clause
func (model *SQLQuery) GroupBy(fields string) *SQLQuery {

	model.pendingGroupBy = append(model.pendingGroupBy, fields)
	return model
}

// OrderBy adds a OrderBy clause
func (model *SQLQuery) OrderBy(fields string, order string) *SQLQuery {

	model.pendingOrderBy = append(model.pendingOrderBy, fields+" "+order)
	return model
}

// Having adds a Having clause
func (model *SQLQuery) Having(field string, cond string) *SQLQuery {

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
func (model *SQLQuery) OrHaving(field string, cond string) *SQLQuery {
	return model.Having(" OR "+field, cond)
}

// Limit adds a Limit clause
func (model *SQLQuery) Limit(limit int) *SQLQuery {
	model.limit = limit
	return model
}

// Offset adds a Offset clause
func (model *SQLQuery) Offset(offset int) *SQLQuery {
	model.offset = offset
	return model
}

// LastQuery returns the last sql query executed
func (model *SQLQuery) LastQuery() string {
	return model.lastQuery
}

// Where adds a Where clause
func (model *SQLQuery) Where(field string, value string) *SQLQuery {

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
func (model *SQLQuery) Select(selectString string) *SQLQuery {

	model.pendingSelects = append(model.pendingSelects, selectString)
	return model
}

// SelectMax adds a SelectMax field to the select
func (model *SQLQuery) SelectMax(selectString string) *SQLQuery {

	return model.Select("MAX(" + selectString + ")")
}

// SelectMin adds a SelectMin field to the select
func (model *SQLQuery) SelectMin(selectString string) *SQLQuery {

	return model.Select("MIN(" + selectString + ")")
}

// SelectAvg adds a SelectAvg field to the select
func (model *SQLQuery) SelectAvg(selectString string) *SQLQuery {

	return model.Select("AVG(" + selectString + ")")
}

// SelectSum adds a SelectSum field to the select
func (model *SQLQuery) SelectSum(selectString string) *SQLQuery {

	return model.Select("Sum(" + selectString + ")")
}

// Find returns the first row with key=id in a struct of ReturnType type
func (model *SQLQuery) Find(id string) (interface{}, error) {

	err := model.
		Limit(1).
		Where(model.key, id).
		executeSelectQuery()

	return model.result[0], err
}

// FindAll returns all the row matching the query in an array of ReturnType type
func (model *SQLQuery) FindAll() ([]interface{}, error) {

	err := model.
		executeSelectQuery()

	return model.result, err
}

// FindAllBy returns all the row matching the fields in an array of ReturnType type
func (model *SQLQuery) FindAllBy(fields map[string]string) ([]interface{}, error) {

	for k, v := range fields {
		model.Where(k, v)
	}

	err := model.
		executeSelectQuery()

	return model.result, err
}

// FindBy returns all the first row matching the on ongoing select in a ReturnType struct
func (model *SQLQuery) FindBy(field string, value string) (interface{}, error) {
	err := model.Where(field, value).executeSelectQuery()
	return model.result[0], err
}

// CountAll returns the number of rows in the table
func (model *SQLQuery) CountAll() (int, error) {

	e := ""
	model.pendingSelects = []string{}
	err := model.
		Select(" count(1) ").
		db.QueryRow(model.composeSelectString()).Scan(&e)

	returnValue, _ := strconv.Atoi(e)
	return returnValue, err
}

// CountBy returns the number of rows in the table with field = value
func (model *SQLQuery) CountBy(field string, value string) (int, error) {

	return model.
		Where(field, value).
		CountAll()
}

// IsUnique returns if field=value is unique in the db
func (model *SQLQuery) IsUnique(field string, value string) (bool, error) {
	count, err := model.
		Where(field, value).
		CountAll()

	if count == 0 {
		return true, err
	}

	return false, err
}

// Insert insert a struct to the db
func (model *SQLQuery) Insert(data interface{}) (bool, error) {

	model.result = []interface{}{data}
	model.executebeforeInsert()

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

	model.executeafterInsert()

	return true, nil
}

// Delete deletes a struct from the db based on key
func (model *SQLQuery) Delete(data interface{}) (bool, error) {

	model.result = []interface{}{data}
	model.executebeforeDelete()

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

	model.executeafterDelete()

	return true, nil

}

//Update sync the data struct with the db according to its model.key field
func (model *SQLQuery) Update(data interface{}) (bool, error) {

	model.result = []interface{}{data}
	model.executebeforeUpdate()

	columnString := []string{}
	var valueString []interface{}
	structPKIndex := -1

	s := reflect.ValueOf(data).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {

		column, dbTagPresent := typeOfT.Field(i).Tag.Lookup("db")

		if dbTagPresent && column != model.key {
			columnString = append(columnString, column+" = ?")
			valueString = append(valueString, s.Field(i).Interface())
		} else if column == model.key {
			structPKIndex = i
		}
	}

	insertStr := "UPDATE " + model.tableName + " SET " +
		strings.Join(columnString, ", ") +
		" WHERE " + model.key + " = ?"

	stmtIns, err := model.db.Prepare(insertStr)

	model.lastQuery = insertStr

	if err != nil {
		return false, err
	}

	result, err := stmtIns.Exec(append(valueString, s.Field(structPKIndex).Interface())...)
	affectedRows, err := result.RowsAffected()

	if err != nil {
		return false, err
	}

	model.executeafterUpdate()

	return affectedRows == 1, nil
}
