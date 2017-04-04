package query

import "database/sql"

type query struct {

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

//Key allow to modify the default id as pk for the table
func (model *query) Key(key string) *query {
	model.key = key
	return model
}

//CreatedField allow to modify model.createdField
func (model *query) CreatedField(createdField string) *query {
	model.createdField = createdField
	return model
}

//ModifiedField allow to modify model.modifiedField
func (model *query) ModifiedField(modifiedField string) *query {
	model.modifiedField = modifiedField
	return model
}

//DeletedField allow to modify model.deletedField
func (model *query) DeletedField(deletedField string) *query {
	model.deletedField = deletedField
	return model
}

//Created allow to modify model.setCreated
func (model *query) Created(created bool) *query {
	model.setCreated = created
	return model
}

//Modified allow to modify model.setModified
func (model *query) Modified(modified bool) *query {
	model.setModified = modified
	return model
}

//SoftDeletes allow to modify model.softDeletes
func (model *query) SoftDeletes(softDeletes bool) *query {
	model.softDeletes = softDeletes
	return model
}

//DateFormat allow to modify model.dateFormat
func (model *query) DateFormat(dateFormat string) *query {
	model.dateFormat = dateFormat
	return model
}

//BeforeInsert sets the BeforeInsert triggers
func (model *SQLQuery) BeforeInsert(triggers []func([]interface{})) *SQLQuery {
	model.beforeInsert = triggers
	return model
}

//AfterInsert sets the AfterInsert triggers
func (model *SQLQuery) AfterInsert(triggers []func([]interface{})) *SQLQuery {
	model.afterInsert = triggers
	return model
}

//BeforeUpdate sets the BeforeUpdate triggers
func (model *SQLQuery) BeforeUpdate(triggers []func([]interface{})) *SQLQuery {
	model.beforeUpdate = triggers
	return model
}

//AfterUpdate sets the AfterUpdate triggers
func (model *SQLQuery) AfterUpdate(triggers []func([]interface{})) *SQLQuery {
	model.afterUpdate = triggers
	return model
}

//BeforeFind sets the BeforeFind triggers
func (model *SQLQuery) BeforeFind(triggers []func([]interface{})) *SQLQuery {
	model.beforeFind = triggers
	return model
}

//AfterFind sets the AfterFind triggers
func (model *SQLQuery) AfterFind(triggers []func([]interface{})) *SQLQuery {
	model.afterFind = triggers
	return model
}

//BeforeDelete sets the BeforeDelete triggers
func (model *SQLQuery) BeforeDelete(triggers []func([]interface{})) *SQLQuery {
	model.beforeDelete = triggers
	return model
}

//AfterDelete sets the AfterDelete triggers
func (model *SQLQuery) AfterDelete(triggers []func([]interface{})) *SQLQuery {
	model.afterDelete = triggers
	return model
}

//executebeforeInsert executes the beforeInsert triggers
func (model *SQLQuery) executebeforeInsert() {
	for index := 0; index < len(model.beforeInsert); index++ {
		model.beforeInsert[index](model.result)
	}
}

//executeafterInsert executes the afterInsert triggers
func (model *SQLQuery) executeafterInsert() {
	for index := 0; index < len(model.afterInsert); index++ {
		model.afterInsert[index](model.result)
	}
}

//executebeforeUpdate executes the beforeUpdate triggers
func (model *SQLQuery) executebeforeUpdate() {
	for index := 0; index < len(model.beforeUpdate); index++ {
		model.beforeUpdate[index](model.result)
	}
}

//executeafterUpdate executes the afterUpdate triggers
func (model *SQLQuery) executeafterUpdate() {
	for index := 0; index < len(model.afterUpdate); index++ {
		model.afterUpdate[index](model.result)
	}
}

//executebeforeFind executes the beforeFind triggers
func (model *SQLQuery) executebeforeFind() {
	for index := 0; index < len(model.beforeFind); index++ {
		model.beforeFind[index](model.result)
	}
}

//executeafterFind executes the afterFind triggers
func (model *SQLQuery) executeafterFind() {
	for index := 0; index < len(model.afterFind); index++ {
		model.afterFind[index](model.result)
	}
}

//executebeforeDelete executes the beforeDelete triggers
func (model *SQLQuery) executebeforeDelete() {
	for index := 0; index < len(model.beforeDelete); index++ {
		model.beforeDelete[index](model.result)
	}
}

//executeafterDelete executes the afterDelete triggers
func (model *SQLQuery) executeafterDelete() {
	for index := 0; index < len(model.afterDelete); index++ {
		model.afterDelete[index](model.result)
	}
}
