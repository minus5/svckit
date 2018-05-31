// Package mssql sadrzi utility funkcije za potrebe DB testova
package mssql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	freetds "github.com/minus5/gofreetds"
	"github.com/stretchr/testify/assert"
)

//InvalidConnStr, je neispravan connect string na bazu
const InvalidConnStr = "user=krivi;password=krivi;server=mssql-unit-test;database=SuperSportUnitTest"

// TestIgracID je id testnog igraca u test bazi, ianic account
const TestIgracID = 3 // ianic

// TestIgracGuid je remember_token testnog igraca u test bazi, ianic account
const TestIgracGuid = "235436360ef4e64add59b56894a488be0f89ff57" // ianic

// TestIgracTesterGuid je remember_token testnog igraca u test bazi, tester account
const TestIgracTesterGuid = "71f3bd95ff4aaa3a6776aa9f509185b046561f78" // tester

// TestKladomatGuid je guid testnog kladomata u test bazi
const TestKladomatGuid = "F5995E98-9A89-424F-BCA5-99FBF6A772CF"

// TestKladomatID je id testnog kladomata u test bazi
const TestKladomatID = 1

// TestKladomatPosID je PosID testnog kladomata u test bazi
const TestKladomatPosID = 1

// TestKladomatSubsidiaryID je SubsidiaryID testnog kladomata u test bazi
const TestKladomatSubsidiaryID = 3

// TestMSSQLUtility sluzi za pristupe bazi, provjer i podesavanja koji nemaju veze sa
// funkcijama i paketima koji se testiraju i zato ima odvojenu konekkciju kako se
// nebi koristila ista konekcija koju korite testirani paketi i funkcije
type TestMSSQLUtility struct {
	pool *freetds.ConnPool
}

// NewTestMSSQLUtility kreira novu instancu i otvara
func NewTestMSSQLUtility(t *testing.T) *TestMSSQLUtility {
	util := &TestMSSQLUtility{}
	return util
}

// DBOpen otvara konekciju za pristup bazi za potrebe paketa i testiranja
func (util *TestMSSQLUtility) DBOpen(t *testing.T) *freetds.ConnPool {
	if nil != util.pool {
		return util.pool
	}
	pool, err := freetds.NewConnPool(util.TestDbConnStr()) // format stringa: "user=...;pwd=...;database=...;host=..."
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	util.pool = pool
	return util.pool
}

// Zatvara konekciju za pristup bazi za potrebe testiranja
func (util *TestMSSQLUtility) DBClose() {
	if nil == util.pool {
		return
	}
	util.pool.Close()
	util.pool = nil
}

//LogDumpStruct kada zelim viditi kako izgleda neki struct, logira u testu
func (util *TestMSSQLUtility) LogDumpStruct(t *testing.T, msg string, value interface{}) {
	buf, err := json.MarshalIndent(value, "  ", "  ")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Logf("%s\n%s\n", msg, buf)
}

// TestDbConnStr je default connect string na test bazu
func (util *TestMSSQLUtility) TestDbConnStr() string {
	s := os.Getenv("GO_SS_TEST_DB")
	if s == "" {
		log.Fatal(` fali connection string za testnu bazu
           nije postavljena env varijabla GO_SS_TEST_DB
           zapisi u ~/.bash_profile (ili slicno) nesto ovako:
           export GO_SS_TEST_DB="user=minus5;password=minus5;server=mssql.s.minus5.hr;database=SuperSportTest_usernameUnit"
           umjesto username treba ici naziv korisnickog racuna (npr. za pero je baza "SuperSportTest_peroUnit")
`)
	}
	return s
}

// ReadFixture cita fixture file
func (util *TestMSSQLUtility) ReadFixture(t *testing.T, fileName string) []byte {
	buf, err := ioutil.ReadFile(fileName)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return buf
}

// TestDbExec izvrsava query na testnoj bazi
func (util *TestMSSQLUtility) TestDbExec(t *testing.T, sql string) {
	util.DBOpen(t)
	util.pool.Do(func(conn *freetds.Conn) error {
		_, err := conn.Exec(sql)
		assert.NoError(t, err)
		return nil
	})
}

// TruncateTable brise sve retke u tablici
func (util *TestMSSQLUtility) TruncateTable(t *testing.T, table string) {
	util.DBOpen(t)
	util.TestDbExec(t, fmt.Sprintf("truncate table %s", table))
}

// RecordsCount vraca broj redaka u trazenoj tablici
func (util *TestMSSQLUtility) RecordsCount(t *testing.T, tableAndWhere string) (int, error) {
	util.DBOpen(t)
	cnt := int32(-1)
	err := util.pool.Do(func(conn *freetds.Conn) error {
		val, err := conn.SelectValue("select count(*) from " + tableAndWhere)
		if err != nil {
			return err
		}
		cnt = val.(int32)
		return nil
	})
	return int(cnt), err
}

// ExecFixture izvrsava ma bazi sadrzaj fixture datoteke, tako da splita po GO
// naredbama i izvrsava kao zasebne query-e
func (util *TestMSSQLUtility) ExecFixture(t *testing.T, name string) {
	f := string(util.ReadFixture(t, name))
	util.DBOpen(t)
	err := util.pool.Do(func(conn *freetds.Conn) error {
		for _, cmd := range strings.Split(f, "GO\n") {
			_, err := conn.Exec(cmd)
			if err != nil {
				return err
			}
		}
		return nil
	})
	assert.NoError(t, err)
}

// SpExists provjerava da li storana procedura postoji u bazi
func (util *TestMSSQLUtility) SpExists(t *testing.T, schemaName, spName string) (bool, error) {
	cnt := 0
	util.DBOpen(t)
	err := util.pool.Do(func(conn *freetds.Conn) error {
		val, err := conn.SelectValue(fmt.Sprintf(`
select count(*) from (
select p.name from sys.procedures p
inner join sys.schemas s on p.schema_id = s.schema_id
where
	s.name = '%s' and p.name = '%s'
) t
`, schemaName, spName))
		if err != nil {
			return err
		}
		cnt = int(val.(int32))
		return nil
	})
	return cnt > 0, err
}

// RunStoredProcedure pokrece storanu proceduru bez parametara na testnoj bazi
func (util *TestMSSQLUtility) RunStoredProcedure(t *testing.T, name string) error {
	util.DBOpen(t)
	return util.pool.Do(func(conn *freetds.Conn) error {
		_, err := conn.Exec("exec " + name)
		return err
	})
}

// RunStoredProcedureWithResult pokrece storanu proceduru s parametarima na testnoj bazi i vraca rezultat
func (util *TestMSSQLUtility) RunStoredProcedureWithResult(t *testing.T, name string, params ...interface{}) (*freetds.SpResult, error) {
	util.DBOpen(t)
	var rez *freetds.SpResult
	err := util.pool.Do(func(conn *freetds.Conn) error {
		var err error
		rez, err = conn.ExecSp(name, params)
		return err
	})
	return rez, err
}

// TestDbExecWithResult izvrsava query sa povratom rezultata
func (util *TestMSSQLUtility) TestDbExecWithResult(t *testing.T, sql string) []*freetds.Result {
	var err error
	var results []*freetds.Result
	util.DBOpen(t)
	util.pool.Do(func(conn *freetds.Conn) error {
		results, err = conn.Exec(sql)
		assert.NoError(t, err)
		return nil
	})
	assert.NotNil(t, results)
	return results
}

// ObrisiPonudu brise kompletnu ponudu u test bazi
func (util *TestMSSQLUtility) ObrisiPonudu(t *testing.T) {
	err := util.RunStoredProcedure(t, "unit_tests.obrisi_ponudu")
	assert.NoError(t, err)
}

// ObrisiListice brise sve listice u test bazi, postavlja rspolozivo tester igraca na 100kn
func (util *TestMSSQLUtility) ObrisiListice(t *testing.T) {
	err := util.RunStoredProcedure(t, "unit_tests.obrisi_listice")
	assert.NoError(t, err)
	util.ResetRaspolozivo(t)
	util.TruncateTable(t, "dbo.SlipEvents")
}

func (util *TestMSSQLUtility) ResetRaspolozivo(t *testing.T) {
	util.TestDbExec(t, fmt.Sprintf("update igraci.igraci set raspolozivo = 100 where remember_token = '%s'", TestIgracGuid))
}

func (util *TestMSSQLUtility) AssertRaspolozivo(t *testing.T, r float64) {
	util.AssertRecordsCount(t, 1, "igraci.igraci where remember_token = '%s' and raspolozivo = %f", TestIgracGuid, r)
}

// AssertRecordsCount prvjerava da li broj redaka u tablici odogovrara ocekivanom
func (util *TestMSSQLUtility) AssertRecordsCount(t *testing.T, expected int, table string, v ...interface{}) {
	table = fmt.Sprintf(table, v...)
	cnt, err := util.RecordsCount(t, table)
	assert.NoError(t, err)
	if !assert.Equal(t, expected, cnt, table) {
		t.Errorf("assertRecordsCount failed expected: %d, actual: %d, table: %s", expected, cnt, table)
		t.Fail()
	}
}
