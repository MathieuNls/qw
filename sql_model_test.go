package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAverage(t *testing.T) {
	s := make([]string, 1)
	s[0] = "root:root@tcp(127.0.0.1:3306)/taxo"
	m, err := NewSQLModel("table", s)

	assert := assert.New(t)
	assert.Nil(err)
}
