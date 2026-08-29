package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	db := newDB()

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Printf("error scanning input: %s\n", err)
			continue
		}

		q := scanner.Text()

		lexer := newLexer(q)
		tokens, err := lexer.lexAll()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		parser := newParser(tokens)
		stmt, err := parser.parse()
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		sqlResult, err := db.execStmt(stmt)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header(sqlResult.headers)
		for _, row := range sqlResult.rows {
			cells := []string{}
			for _, cell := range row {
				cells = append(cells, cell.String())
			}
			table.Append(cells)
		}
		table.Render()

		fmt.Print("> ")
	}
}
