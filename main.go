package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

const (
	migrationsTableExists = `SELECT EXISTS (SELECT 1 FROM information_schema.tables 
		WHERE table_name = 'pgmigrate')`
	createMigrationsTable = `CREATE TABLE pgmigrate(id SERIAL, migration TEXT, 
		created_at TIMESTAMP DEFAULT now());`
	latestMigration = `SELECT migration FROM pgmigrate ORDER BY migration DESC`
	addMigration    = `INSERT INTO pgmigrate(migration) VALUES($1)`
	removeMigration = `DELETE FROM pgmigrate WHERE migration=$1`
)

const (
	upCommand   = "up"
	downCommand = "down"
)

var (
	dsn string
	dir string
	cmd string
)

func main() {
	validateInput()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	validateDB(db)

	lastMigration := getLastMigration(db)

	if cmd == downCommand {
		down(db, dir, lastMigration)
		os.Exit(0)
	}

	up(db, dir, lastMigration)
	os.Exit(0)
}

func validateInput() {
	flag.StringVar(&dsn, "dsn",
		"postgres://postgres:@localhost/postgres?sslmode=disable",
		"The DSN to use to connect to the database")
	flag.StringVar(&dir, "dir", "sql", "The path to the directory of migration files")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pgmigrate [OPTIONS] COMMAND")
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmd = flag.Args()[0]
	if cmd != upCommand && cmd != downCommand {
		fmt.Fprintf(os.Stderr, "valid commands are '%s' and '%s'\n", upCommand, downCommand)
		os.Exit(1)
	}

	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "%s does not exist or isn't a directory: %s\n", dir, err)
		os.Exit(1)
	}
}

func validateDB(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not ping database: %s\n", err)
		os.Exit(1)
	}

	var exists bool
	err = db.QueryRow(migrationsTableExists).Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
		os.Exit(1)
	}

	if !exists {
		_, err = db.Exec(createMigrationsTable)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create migrations table: %s\n", err)
			os.Exit(1)
		}
	}
}

func getLastMigration(db *sql.DB) string {
	var lastMigration string
	err := db.QueryRow(latestMigration).Scan(&lastMigration)
	if err != nil && err != sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
		os.Exit(1)
	}
	return lastMigration
}

func up(db *sql.DB, dir, lastMigration string) {
	migrations := make([]string, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading the migrations: %s\n", err)
		os.Exit(1)
	}
	for _, f := range files {
		n := f.Name()
		if strings.Contains(n, "up.sql") {
			if strings.TrimSuffix(n, ".up.sql") > lastMigration {
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
		_, err = db.Exec(string(addMigration), mn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error saving migration metadata: %s\n", err)
			fmt.Fprintln(os.Stderr, "[WARNING] database in inconsistent state")
			os.Exit(1)
		}
	}
	fmt.Fprintf(os.Stdout, "Applied %d migrations\n", len(migrations))
}

func down(db *sql.DB, dir, lastMigration string) {
	if lastMigration == "" {
		os.Exit(0)
	}
	fn := filepath.Join(dir, lastMigration+".down.sql")
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
	_, err = db.Exec(string(removeMigration), lastMigration)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error querying the database: %s\n", err)
		fmt.Fprintln(os.Stderr, "[WARNING] database in inconsistent state")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Migration %s successfully removed\n", lastMigration)
}
