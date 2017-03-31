package qw

import (
	"database/sql"
	"testing"

	"errors"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

//Mock database connection
type CnxMock struct {
	Db   *sql.DB
	Mock sqlmock.Sqlmock
}

//Mock implementation of the CnxOpener type
func (m *CnxMock) OpenCnx(dbCons []string) (*sql.DB, error) {

	db, mock, err := sqlmock.New()
	m.Db = db
	m.Mock = mock
	return db, err
}

func TestNew(t *testing.T) {

	s := []string{
		"mock:mock@mock(127.0.0.1:3306)/mock",
	}
	_, err := NewSQLQuery("bugs", s, new(CnxMock))

	assert := assert.New(t)
	assert.Nil(err)
}

func TestCleanup(t *testing.T) {

	s := []string{
		"mock:mock@mock(127.0.0.1:3306)/mock",
	}
	m, err := NewSQLQuery("bugs", s, new(CnxMock))

	assert := assert.New(t)
	assert.Nil(err)

	//Assign random values
	m.pendingSelects = []string{"mock"}
	m.pendingWheres = []string{"mock"}
	m.pendingJoins = []string{"mock"}
	m.pendingUnions = []string{"mock"}
	m.pendingGroupBy = []string{"mock"}
	m.pendingOrderBy = []string{"mock"}
	m.limit = 25
	m.offset = 25

	m.cleanup(errors.New("mock"))

	assert.Empty(m.pendingSelects, "should be empty")
	assert.Empty(m.pendingWheres, "should be empty")
	assert.Empty(m.pendingJoins, "should be empty")
	assert.Empty(m.pendingUnions, "should be empty")
	assert.Empty(m.pendingGroupBy, "should be empty")
	assert.Empty(m.pendingOrderBy, "should be empty")
	assert.Equal(-1, m.limit, "should be -1")
	assert.Equal(-1, m.offset, "should be -1")
	assert.Equal("mock", m.lastError.Error(), "should be `mock`")
}

func TestComposeSelectString(t *testing.T) {
	s := []string{
		"mock:mock@mock(127.0.0.1:3306)/mock",
	}
	m, err := NewSQLQuery("mock", s, new(CnxMock))

	selectStr := m.
		Select("a, b").
		Select("c").
		SelectAvg("d").
		SelectMax("e").
		SelectMin("f").
		SelectSum("g").
		GroupBy("h, i").
		OrderBy("j", "ASC").
		OrderBy("k", "DESC").
		Like("l", "l").
		NotLike("m", "m").
		OrLike("n", "n").
		OrNotLike("o", "o").
		WhereIn("p", []string{"p", "p"}).
		WhereNotIn("q", []string{"q", "q"}).
		OrWhereIn("s", []string{"s", "s"}).
		OrWhereNotIn("t", []string{"t", "t"}).
		Having("u", "count(u) > 1").
		OrHaving("v", "count(v) > 1").
		Limit(28).
		Offset(42).
		Join("w", "w.a = mock.a", "").
		Join("x", "x.a = mock.a", "left").
		Join("y", "y.a = mock.a", "right").
		Where("a >=", "2").
		Where("a <=", "3").
		Where("a >", "3").
		Where("a <", "3").
		OrWhere("b <=", "3").
		OrWhere("b >", "3").
		OrWhere("b <", "3").
		OrWhere("b !=", "3").
		OrWhere("b <>", "3").
		composeSelectString()

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal("SELECT a, b, c, AVG(d), MAX(e), MIN(f), Sum(g) FROM mock JOIN  w ON w.a = mock.a left JOIN  x ON x.a = mock.a right JOIN  y ON y.a = mock.a WHERE l LIKE 'l'  AND m NOT LIKE 'm'  OR n LIKE 'n'  OR o NOT LIKE 'o'  AND p IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  AND q NOT IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  OR s IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  OR t NOT IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  AND a >= 2  AND a <= 3  AND a > 3  AND a < 3  OR b <= 3  OR b > 3  OR b < 3  OR b != 3  OR b <> 3 GROUP BY h, i HAVING u count(u) > 1  OR v count(v) > 1 ORDER BY h, i LIMIT 28 OFFSET  42", selectStr)
	assert.Equal("SELECT a, b, c, AVG(d), MAX(e), MIN(f), Sum(g) FROM mock JOIN  w ON w.a = mock.a left JOIN  x ON x.a = mock.a right JOIN  y ON y.a = mock.a WHERE l LIKE 'l'  AND m NOT LIKE 'm'  OR n LIKE 'n'  OR o NOT LIKE 'o'  AND p IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  AND q NOT IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  OR s IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  OR t NOT IN 'a, b, c, AVG(d), MAX(e), MIN(f), Sum(g)'  AND a >= 2  AND a <= 3  AND a > 3  AND a < 3  OR b <= 3  OR b > 3  OR b < 3  OR b != 3  OR b <> 3 GROUP BY h, i HAVING u count(u) > 1  OR v count(v) > 1 ORDER BY h, i LIMIT 28 OFFSET  42", m.LastQuery())

}

func TestReflectResult(t *testing.T) {

	s := []string{
		"mock:mock@mock(127.0.0.1:3306)/mock",
	}

	values := []sql.RawBytes{
		[]byte("1"),
		[]byte("test"),
		[]byte("1.2"),
	}

	type T struct {
		ID           int     `db:"ID"`
		Name         string  `db:"NAME"`
		AnotherFloat float64 `db:"TEST"`
	}

	columns := []string{"ID", "NAME", "TEST"}

	m, err := NewSQLQuery("bugs", s, new(CnxMock))
	m.returnType = new(T)

	reflectedStruct := m.reflectResult(values, columns)

	assert := assert.New(t)
	assert.Nil(err)

	assert.Equal(1, reflectedStruct.(T).ID)
	assert.Equal("test", reflectedStruct.(T).Name)
	assert.Equal(1.2, reflectedStruct.(T).AnotherFloat)
}

//Need mock
// func TestUpdate(t *testing.T) {

// 	type Bug struct {
// 		ID    int    `db:"INTERNAL_ID"`
// 		ExtID string `db:"EXTERNAL_ID"`
// 	}

// 	s := []string{
// 		"root:root@tcp(127.0.0.1:3306)/taxo",
// 	}

// 	b := new(Bug)
// 	b.ExtID = "aaaaa"
// 	b.ID = 5

// 	model, err := NewSQLModel("bugs", s, new(MySQLCnxOpenner))

// 	model.AfterUpdate(
// 		[]func([]interface{}){
// 			func(rows []interface{}) {

// 				for index := 0; index < len(rows); index++ {
// 					fmt.Println(rows[0])
// 					rows[0].(*Bug).ExtID = "awdawdawdawdawd"
// 					fmt.Println(rows[0])
// 				}
// 			},
// 		},
// 	)

// 	r, err := model.
// 		Key("INTERNAL_ID").
// 		Update(b)

// 	fmt.Println(r)
// 	fmt.Println(err)

// }
