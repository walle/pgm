package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: pgmigrate DSN COMMAND DIRECTORY")
		os.Exit(1)
	}

	dsn := os.Args[1]

	cmd := os.Args[2]
	if cmd != "up" && cmd != "down" {
		fmt.Fprintln(os.Stderr, "valid commands are 'up' and 'down'")
		os.Exit(1)
	}

	dir := os.Args[3]
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "%s does not exist or isn't a directory: %s\n", dir, err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not ping database: %s\n", err)
		os.Exit(1)
	}

	var exists bool
	eq := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'pgmigrate')`
	err = db.QueryRow(eq).Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
		os.Exit(1)
	}

	if !exists {
		cq := `CREATE TABLE pgmigrate(id SERIAL, migration TEXT, 
		created_at TIMESTAMP DEFAULT now());`
		_, err = db.Exec(cq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create migrations table: %s\n", err)
			os.Exit(1)
		}
	}

	var latestMigration string
	lmq := `SELECT migration FROM pgmigrate ORDER BY migration DESC`
	err = db.QueryRow(lmq).Scan(&latestMigration)
	if err != nil && err != sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
		os.Exit(1)
	}

	if cmd == "up" {
		migrations := make([]string, 0)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading the migrations: %s\n", err)
			os.Exit(1)
		}
		for _, f := range files {
			n := f.Name()
			if strings.Contains(n, "up.sql") {
				if strings.TrimSuffix(n, ".up.sql") > latestMigration {
					migrations = append(migrations, n)
				}
			}
		}
		for _, m := range migrations {
			fq, err := ioutil.ReadFile(filepath.Join(dir, m))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading the migration: %s\n", err)
				os.Exit(1)
			}
			_, err = db.Exec(string(fq))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
				os.Exit(1)
			}
			mn := strings.TrimSuffix(m, ".up.sql")
			cmq := `INSERT INTO pgmigrate(migration) VALUES($1)`
			_, err = db.Exec(string(cmq), mn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error saving migration metadata: %s\n", err)
				fmt.Fprintln(os.Stderr, "[WARNING] database in inconsistent state")
				os.Exit(1)
			}
		}
	} else if cmd == "down" {
		if latestMigration == "" {
			os.Exit(0)
		}
		fn := filepath.Join(dir, latestMigration+".down.sql")
		if _, err := os.Stat(fn); err != nil {
			fmt.Fprintf(os.Stderr, "%s does not exist: %s\n", fn, err)
			os.Exit(1)
		}
		fq, err := ioutil.ReadFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading the migration: %s\n", err)
			os.Exit(1)
		}
		_, err = db.Exec(string(fq))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
			os.Exit(1)
		}
		rmq := `DELETE FROM pgmigrate WHERE migration=$1`
		_, err = db.Exec(string(rmq), latestMigration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
			fmt.Fprintln(os.Stderr, "[WARNING] database in inconsistent state")
			os.Exit(1)
		}
		os.Exit(0)
	}
}
