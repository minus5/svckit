package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPostgresUtility sluzi za pristup bazi
type TestPostgresUtility struct {
	conn *sql.DB
}

// NewTestPostgresUtility kreira novu instancu
func NewTestPostgresUtility(t *testing.T) *TestPostgresUtility {
	util := &TestPostgresUtility{}
	return util
}

// DBOpen otvara konekciju za pristup bazu
func (util *TestPostgresUtility) DBOpen(t *testing.T) *sql.DB {
	if util.conn != nil {
		return util.conn
	}
	connStr := util.TestDbConnStr()
	db, err := sql.Open("postgres", connStr)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	util.conn = db
	return util.conn

}

// DBClose zatvara konekciju za pristup bazi
func (util *TestPostgresUtility) DBClose(t *testing.T) {
	if util.conn == nil {
		return
	}
	util.conn.Close()
	util.conn = nil
}

// TestDbConnStr dohvaca env varijablu POSTGRES_TEST_DB sa default stringom za bazu
func (util *TestPostgresUtility) TestDbConnStr() string {
	s := os.Getenv("POSTGRES_TEST_DB")
	if s == "" {
		log.Fatal("Nedostaje env varijabla POSTGRES_TEST_DB")
	}
	return s
}

// TestDbExec izvrsava query na bazi
func (util *TestPostgresUtility) TestDbExec(t *testing.T, sql string) {
	util.DBOpen(t)
	_, err := util.conn.Exec(sql)
	assert.NoError(t, err)
}

// TestDbExecWithResult izvrsava query sa povratom rezultata
func (util *TestPostgresUtility) TestDbExecWithResult(t *testing.T, sql string) (*sql.Rows, error) {
	util.DBOpen(t)
	rows, err := util.conn.Query(sql)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// TruncateTable brise sve retke iz tablice
func (util *TestPostgresUtility) TruncateTable(t *testing.T, table string) {
	util.DBOpen(t)
	util.TestDbExec(t, fmt.Sprintf("truncate table %s restart identity", table))
}

// RecordsCount vraca broj redaka u tablici
func (util *TestPostgresUtility) RecordsCount(t *testing.T, table string) (int, error) {
	util.DBOpen(t)

	var count int
	row := util.conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	err := row.Scan(&count)

	if err != nil {
		return -1, err
	}

	return count, nil
}

func (util *TestPostgresUtility) ReadFixture(t *testing.T, fileName string) []byte {
	buf, err := ioutil.ReadFile(fileName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return buf
}

func (util *TestPostgresUtility) ExecFixture(t *testing.T, name string) {
	f := string(util.ReadFixture(t, name))
	util.DBOpen(t)
	for _, cmd := range strings.Split(f, "GO\n") {
		util.TestDbExec(t, cmd)
	}
}

// SpExists provjerava postoji li procedura
func (util *TestPostgresUtility) SpExists(t *testing.T, schemaName, spName string) (bool, error) {
	util.DBOpen(t)
	var exists bool
	row := util.conn.QueryRow(fmt.Sprintf(`
		SELECT EXISTS (
			SELECT *
				FROM pg_catalog.pg_proc
			JOIN pg_namespace ON pg_catalog.pg_proc.pronamespace = pg_namespace.oid
			WHERE proname = '%s'
			AND pg_namespace.nspname = '%s'
		)
`, spName, schemaName))
	err := row.Scan(&exists)
	if err != nil {
		return exists, err
	}
	return exists, nil
}

// RunStoredProcedure pokrece proceduru bez parametara
func (util *TestPostgresUtility) RunStoredProcedure(t *testing.T, name string) error {
	util.DBOpen(t)
	_, err := util.conn.Exec(fmt.Sprintf("SELECT %s()", name))
	return err
}

func (util *TestPostgresUtility) AssertRecordsCount(t *testing.T, expected int, table string, v ...interface{}) {
	table = fmt.Sprintf(table, v...)
	cnt, err := util.RecordsCount(t, table)
	assert.NoError(t, err)
	if !assert.Equal(t, expected, cnt, table) {
		t.Errorf("AssertRecordsCount failed. Expected: %d, actual: %d, table: %s", expected, cnt, table)
		t.Fail()
	}
}
