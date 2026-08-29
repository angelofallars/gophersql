package main

import (
	"errors"
	"fmt"
	"slices"
)

type db struct {
	tables map[string]Table
	values map[string][]Row
}

func newDB() db {
	return db{
		tables: make(map[string]Table),
		values: make(map[string][]Row),
	}
}

func (d *db) execStmt(stmt any) (sqlResult, error) {
	switch stmt := stmt.(type) {
	case StmtCreateTable:
		if _, ok := d.tables[stmt.tableName]; ok {
			return sqlResult{}, fmt.Errorf("table '%s' already exists", stmt.tableName)
		}
		table := Table{
			name:    stmt.tableName,
			columns: stmt.columns,
		}

		d.tables[stmt.tableName] = table

		return sqlResult{headers: table.columnNames()}, nil
	case StmtInsert:
		table, ok := d.tables[stmt.tableName]
		if !ok {
			return sqlResult{}, fmt.Errorf("table '%s' does not exist", stmt.tableName)
		}
		if len(stmt.columns) != len(stmt.values) {
			return sqlResult{}, fmt.Errorf("inconsistent amount of inserted values")
		}

		row := newRow(table)

		for i := range len(stmt.columns) {
			columnName := stmt.columns[i]
			value := stmt.values[i]

			idx := -1
			var column Column
			for j, col := range table.columns {
				if columnName == col.name {
					idx = j
					column = col
					break
				}
			}
			if idx == -1 {
				return sqlResult{}, fmt.Errorf("column '%s' does not exist", columnName)
			}

			if value.kind != column.kind {
				return sqlResult{}, fmt.Errorf("expected type '%s' for column '%s', got '%s'", column.kind, columnName, value.kind)
			}

			row[idx] = value
		}

		d.values[table.name] = append(d.values[table.name], row)

		return sqlResult{headers: table.columnNames(), rows: []Row{row}}, nil
	case StmtSelect:
		table, ok := d.tables[stmt.tableName]
		if !ok {
			return sqlResult{}, fmt.Errorf("table '%s' does not exist", stmt.tableName)
		}

		tableColumnNames := table.columnNames()
		idxsToFetch := []int{}
		for _, column := range stmt.columns {
			idx := slices.Index(tableColumnNames, column)
			if idx == -1 {
				return sqlResult{}, fmt.Errorf("column '%s' does not exist", column)
			}
			idxsToFetch = append(idxsToFetch, idx)
		}

		resultRows := []Row{}
		for _, row := range d.values[table.name] {
			resultRow := Row{}
			for _, idx := range idxsToFetch {
				resultRow = append(resultRow, row[idx])
			}
			resultRows = append(resultRows, resultRow)
		}

		return sqlResult{headers: stmt.columns, rows: resultRows}, nil
	default:
		return sqlResult{}, errors.New("unrecognized statement")
	}
}

type Table struct {
	name    string
	columns []Column
}

func (t *Table) columnNames() []string {
	names := []string{}
	for _, column := range t.columns {
		names = append(names, column.name)
	}
	return names
}

type Row []Value

func newRow(t Table) Row {
	r := Row{}
	for _, column := range t.columns {
		switch column.kind {
		case ValueKindInt:
			r = append(r, Value{kind: ValueKindInt})
		case ValueKindText:
			r = append(r, Value{kind: ValueKindText})
		default:
			panic("unrecognized value kind")
		}
	}

	return r
}

type sqlResult struct {
	headers []string
	rows    []Row
}
