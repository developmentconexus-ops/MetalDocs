package application

import (
	"database/sql/driver"
	"io"
)

type schedulerEmptyRows struct{}

func (schedulerEmptyRows) Columns() []string         { return nil }
func (schedulerEmptyRows) Close() error              { return nil }
func (schedulerEmptyRows) Next([]driver.Value) error { return io.EOF }
