package postgresdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/sectiontrace"

	_ "github.com/lib/pq"
)

type SchemaUpgrade struct {
	Next int
	Sql  []string
}

type Schema struct {
	Name           string
	Upgrades       map[int]SchemaUpgrade
	CurrentVersion int
}

type sectionmaker struct {
	mu       sync.Mutex
	sections map[string]sectiontrace.Section
}

var sections = &sectionmaker{}

func (s *sectionmaker) Get(name string) sectiontrace.Section {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sections == nil {
		s.sections = map[string]sectiontrace.Section{}
	}

	name = fmt.Sprintf("postgresdb.%s", name)

	sec, ok := s.sections[name]
	if !ok {
		sec = sectiontrace.New(name)
		s.sections[name] = sec
	}

	return sec
}

func SequentialUpgrades(upgrades ...[]string) map[int]SchemaUpgrade {
	m := map[int]SchemaUpgrade{}
	for i, upgrade := range upgrades {
		m[i] = SchemaUpgrade{Sql: upgrade}
	}
	return m
}

type Database struct {
	schema *Schema
	db     *sql.DB
}

const (
	schemaTableName = "___orcschema"
)

var (
	defaultTxOpts = &sql.TxOptions{Isolation: sql.LevelSerializable}
)

type Queryer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func OpenRawDB(ctx context.Context, rawConnstring, password string) (*sql.DB, error) {
	parsed, err := url.Parse(rawConnstring)
	_, hasPassword := parsed.User.Password()
	if hasPassword {
		return nil, fmt.Errorf("postgres connection string should not contain password")
	}

	if parsed.User == nil {
		parsed.User = &url.Userinfo{}
	}
	parsed.User = url.UserPassword(parsed.User.Username(), password)

	secretConnstring := parsed.String()

	db, err := sql.Open("postgres", secretConnstring)
	if err != nil {
		return nil, fmt.Errorf("Unable to open database: %v", err)
	}

	return db, nil
}

func (s *Schema) Open(ctx context.Context, rawConnstring, password string) (*Database, error) {
	db, err := OpenRawDB(ctx, rawConnstring, password)
	if err != nil {
		return nil, err
	}

	rv := &Database{
		schema: s,
		db:     db,
	}

	if err := rv.startup(ctx); err != nil {
		return nil, fmt.Errorf("Unable to open database: %v", err)
	}

	return rv, nil
}

func createMetatable(ctx context.Context, q Queryer, schemaName string) error {
	sqlquery1 := `CREATE TABLE ___orcschema (
		name TEXT NOT NULL,
		version INTEGER NOT NULL,
		meta_version INTEGER NOT NULL
	);
	`
	sqlquery2 := `INSERT INTO ___orcschema (name, version, meta_version) VALUES ($1, $2, $3);`

	initialVersion := int(0)
	initialMetaVersion := int(1)

	if _, err := q.ExecContext(ctx, sqlquery1); err != nil {
		return fmt.Errorf("error creating metatable: %v", err)
	}

	if _, err := q.ExecContext(ctx, sqlquery2, schemaName, initialVersion, initialMetaVersion); err != nil {
		return fmt.Errorf("error creating metatable: %v", err)
	}

	return nil
}

func doesMetatableExist(ctx context.Context, q Queryer) (bool, error) {
	sqlquery := `SELECT table_name AS name FROM information_schema.tables WHERE table_schema = 'public';`
	rows, err := q.QueryContext(ctx, sqlquery)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var sawMetatable bool
	var sawOtherTables []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return false, err
		}

		if name != schemaTableName {
			sawOtherTables = append(sawOtherTables, name)
		} else {
			sawMetatable = true
		}
	}

	switch {
	case sawMetatable:
		return true, nil
	case len(sawOtherTables) > 0:
		return false, fmt.Errorf("Database expectation mismatch: did not see %q but saw other tables: %v", schemaTableName, sawOtherTables)
	default:
		return false, nil
	}
}

func (d *Database) startup(ctx context.Context) error {
	if d.schema.Name == "" {
		return fmt.Errorf("Invalid schema: missing name")
	}

	exists, err := doesMetatableExist(ctx, d.db)
	if err != nil {
		return err
	}

	if !exists {
		if err := d.runInTransaction(ctx, defaultTxOpts, func(tx *sql.Tx) error {
			return createMetatable(ctx, tx, d.schema.Name)
		}); err != nil {
			return err
		}
	}

	name, version, err := getSchemaVersion(ctx, d.db)
	if err != nil {
		return err
	}

	if name != d.schema.Name {
		return fmt.Errorf("Database schema mismatch (got %q want %q)", name, d.schema.Name)
	}

	if err := d.performUpgrades(ctx); err != nil {
		return err
	}

	if d.schema.CurrentVersion != 0 {
		_, upgradedVersion, err := getSchemaVersion(ctx, d.db)
		if err != nil {
			return err
		}
		if upgradedVersion != d.schema.CurrentVersion {
			return fmt.Errorf("Database version expectation failure (got %d => %d want %d)", version, upgradedVersion, d.schema.CurrentVersion)
		}
	}

	return nil
}

func Transactor(transactionName string) func(context.Context, *Database, func(context.Context, *sql.Tx) error) error {
	tracer := sections.Get(transactionName)
	return func(ctx context.Context, db *Database, callback func(ctx context.Context, tx *sql.Tx) error) error {
		newCtx, sec := tracer.Begin(ctx)
		err := db.runInTransaction(ctx, defaultTxOpts, func(tx *sql.Tx) error {
			err := callback(newCtx, tx)
			return err
		})
		sec.End(err)
		return err
	}
}

func (d *Database) runInTransaction(ctx context.Context, opts *sql.TxOptions, callback func(tx *sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	if err := callback(tx); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("%v, then rollback error: %v", err, rollbackErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("Commit error: %v, then rollback error: %v", err, rollbackErr)
		}
		return err
	}

	return nil
}

func getSchemaVersion(ctx context.Context, q Queryer) (string, int, error) {
	sqlquery := `SELECT name, version FROM ___orcschema;`
	var name string
	var version int
	if err := q.QueryRowContext(ctx, sqlquery).Scan(&name, &version); err != nil {
		return "", 0, err
	}

	return name, version, nil
}

func execStatement(ctx context.Context, q Queryer, stmt string, params ...interface{}) error {
	_, err := q.ExecContext(ctx, stmt, params...)
	return err
}

func execStatements(ctx context.Context, q Queryer, stmts []string) error {
	for _, stmt := range stmts {
		logrus.Infof("executing statement: %q", stmt)
		if err := execStatement(ctx, q, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (d *Database) applyUpgrade(ctx context.Context, sqltexts []string, oldVer, newVer int) error {
	opts := &sql.TxOptions{Isolation: sql.LevelSerializable}
	err := d.runInTransaction(ctx, opts, func(tx *sql.Tx) error {
		_, version, err := getSchemaVersion(ctx, tx)
		if err != nil {
			return err
		}

		if version != oldVer {
			return fmt.Errorf("Version expectation mismatch during upgrade: upgrading %d => %d, yet version was %d", oldVer, newVer, version)
		}

		if err := execStatements(ctx, tx, sqltexts); err != nil {
			return err
		}

		_, version, err = getSchemaVersion(ctx, tx)
		if err != nil {
			return err
		}

		if version != oldVer {
			return fmt.Errorf("Version expectation mismatch during upgrade: upgrading %d => %d, yet version was %d after script ran", oldVer, newVer, version)
		}

		updateStmt := `UPDATE ___orcschema SET version = $1 ;`
		if err := execStatement(ctx, tx, updateStmt, newVer); err != nil {
			return err
		}

		_, version, err = getSchemaVersion(ctx, tx)
		if err != nil {
			return err
		}

		if version != newVer {
			return fmt.Errorf("Version expectation mismatch during upgrade: upgrading %d => %d, yet version was %d after version should have been updated", oldVer, newVer, version)
		}

		return nil
	})
	logrus.Infof("Performed database upgrade: %d => %d: err: %v", oldVer, newVer, err)
	return err
}

func (d *Database) performUpgrades(ctx context.Context) error {
	if d.schema.Upgrades == nil {
		return nil
	}

	for {
		_, version, err := getSchemaVersion(ctx, d.db)
		if err != nil {
			return err
		}

		if d.schema.Upgrades != nil {
			upgrade, ok := d.schema.Upgrades[version]
			if !ok {
				return nil
			}

			nextVer := version + 1
			if upgrade.Next != 0 {
				nextVer = upgrade.Next
			}

			if nextVer <= version {
				return fmt.Errorf("Invalid update (%d => %d): version must increase", version, nextVer)
			}

			if err := d.applyUpgrade(ctx, upgrade.Sql, version, nextVer); err != nil {
				return err
			}
		}
	}
}

func (d *Database) Close() error {
	var err error
	if d.db != nil {
		err = d.db.Close()
	}
	d.db = nil
	return err
}

func fromArgmap(paramNames []string, argmap map[string]interface{}) ([]interface{}, error) {
	args := make([]interface{}, len(paramNames))

	for i, name := range paramNames {
		value, ok := argmap[name]
		if !ok {
			return nil, fmt.Errorf("missing parameter %q", name)
		}
		args[i] = value
	}

	if argmap == nil {
		if len(paramNames) > 0 {
			return nil, fmt.Errorf("want %d parameters got 0 (nil)", len(paramNames))
		}
	} else {
		if len(paramNames) != len(argmap) {
			var ks []string
			for k := range argmap {
				ks = append(ks, k)
			}
			return nil, fmt.Errorf("want %d parameters got %d (%v)", len(paramNames), len(argmap), ks)
		}
	}

	return args, nil
}

type QueryFailed struct {
	QueryName string
	Err       error
}

func (q QueryFailed) Error() string {
	return fmt.Sprintf("query %q failed: %v", q.QueryName, q.Err)
}

type PreparedExec struct {
	section    sectiontrace.Section
	stmt       *sql.Stmt
	queryName  string
	paramNames []string
}

func (p *PreparedExec) ExecWithResult(ctx context.Context, tx *sql.Tx, argmap map[string]interface{}) (sql.Result, error) {
	var rv sql.Result
	err := p.section.Do(ctx, func(ctx context.Context) error {
		args, err := fromArgmap(p.paramNames, argmap)
		if err != nil {
			return QueryFailed{p.queryName, err}
		}

		result, err := tx.Stmt(p.stmt).ExecContext(ctx, args...)
		if err != nil {
			return QueryFailed{p.queryName, err}
		}
		rv = result
		return nil
	})
	return rv, err
}

func (p *PreparedExec) Exec(ctx context.Context, tx *sql.Tx, argmap map[string]interface{}) error {
	_, err := p.ExecWithResult(ctx, tx, argmap)
	return err
}

func (d *Database) PrepareInsertExec(outErr *error, tableName string, fieldNames []string) *PreparedExec {
	sqlText := "INSERT INTO " + tableName + "("
	sqlText += strings.Join(fieldNames, ",")
	sqlText += ") VALUES ("
	for i := range fieldNames {
		if i > 0 {
			sqlText += ","
		}
		sqlText += fmt.Sprintf("$%d", i+1)
	}
	sqlText += ");"
	queryName := fmt.Sprintf("insert-%s-(%s)", tableName, strings.Join(fieldNames, ","))
	return d.PrepareExec(outErr, queryName, sqlText, fieldNames...)
}

func (d *Database) PrepareExec(outErr *error, queryName, querySQL string, paramNames ...string) *PreparedExec {
	if *outErr != nil {
		return nil
	}

	stmt, err := d.db.Prepare(querySQL)
	if err != nil {
		*outErr = fmt.Errorf("Failed to prepare query %q: %v"+`
Query was: """
%s
"""`, queryName, err, querySQL)
		return nil
	}

	sec := sections.Get(queryName)
	return &PreparedExec{
		section:    sec,
		queryName:  queryName,
		paramNames: paramNames,
		stmt:       stmt,
	}
}

type PreparedQuery struct {
	section    sectiontrace.Section
	stmt       *sql.Stmt
	queryName  string
	paramNames []string
}

func makeQueryDest(names []string, dest interface{}) ([]interface{}, error) {
	if len(names) == 0 {
		return nil, nil
	}

	nameMap := map[string]int{}
	for i, name := range names {
		name = strings.Replace(name, "_", "", -1)
		name = strings.ToLower(name)
		nameMap[name] = i
	}

	destptrs := make([]interface{}, len(names))

	structValue := reflect.ValueOf(dest).Elem()
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		name := strings.ToLower(field.Name)
		index, ok := nameMap[name]
		if !ok {
			index, ok = nameMap[field.Tag.Get("sql")]
		}
		if !ok {
			return nil, fmt.Errorf("Struct field %q does not match any field (%v)", field.Name, names)
		}
		if index < 0 || index >= len(destptrs) {
			return nil, fmt.Errorf("Struct field %q at index %d out of range (%d)", field.Name, i, len(destptrs))
		}
		if destptrs[index] != nil {
			return nil, fmt.Errorf("Struct field %q is duplicate", field.Name)
		}
		destptrs[index] = structValue.Field(i).Addr().Interface()
	}

	return destptrs, nil
}

func (p *PreparedQuery) Query(ctx context.Context, tx *sql.Tx, argmap map[string]interface{}, dest interface{}, onrow func() (bool, error)) error {
	return p.section.Do(ctx, func(ctx context.Context) error {
		args, err := fromArgmap(p.paramNames, argmap)
		if err != nil {
			return QueryFailed{p.queryName, err}
		}

		rows, err := tx.Stmt(p.stmt).QueryContext(ctx, args...)
		if err != nil {
			return QueryFailed{p.queryName, err}
		}
		defer rows.Close()

		names, err := rows.Columns()
		if err != nil {
			return err
		}

		destparams, err := makeQueryDest(names, dest)
		if err != nil {
			return err
		}

		for rows.Next() {
			if err := rows.Scan(destparams...); err != nil {
				return err
			}

			cont := true
			var err error

			if onrow != nil {
				cont, err = onrow()
				if err != nil {
					return err
				}
			}

			if !cont {
				break
			}
		}

		return nil
	})
}

func (d *Database) PrepareQuery(outErr *error, queryName, querySQL string, paramNames ...string) *PreparedQuery {
	if *outErr != nil {
		return nil
	}

	stmt, err := d.db.Prepare(querySQL)
	if err != nil {
		*outErr = fmt.Errorf("Failed to prepare query %q: %v", queryName, err)
		return nil
	}

	sec := sections.Get(queryName)
	return &PreparedQuery{
		section:    sec,
		queryName:  queryName,
		stmt:       stmt,
		paramNames: paramNames,
	}
}
