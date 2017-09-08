package keys

import (
	"flag"
	"fmt"
	"io/ioutil"
	"pkg/testu"
	"testing"

	simplejson "github.com/minus5/go-simplejson"
	"github.com/stretchr/testify/assert"
)

var saveFixtures bool

func init() {
	flag.BoolVar(&saveFixtures, "save-fixtures", false, "snimi fixture umjesto testiranja spram njih")
	flag.Parse()
}

func TestExtender(t *testing.T) {
	first := `
{
  "i": 1,
  "n": "Soccer",
  "o": {
     "1": {
        "j": 1,
        "d": "Dinamo",
        "g": "Hajduk"
        }
     }
  }
}`

	j, err := simplejson.NewJson([]byte(first))
	assert.Nil(t, err)
	assert.NotNil(t, j)

	j.Set("p", "jozo")
	j.Set("p", j.Get("p"))

	km := map[string]string{
		"i": "id",
		"n": "naziv",
		"o": "dogadjaji",
		"d": "domacin",
		"g": "gost",
		"p": "pero",
	}

	e := NewExtender(j)
	out := e.ExtendWith(km)

	expected := `{
  "dogadjaji": {
    "1": {
      "domacin": "Dinamo",
      "gost": "Hajduk",
      "j": 1
    }
  },
  "id": 1,
  "naziv": "Soccer",
  "pero": "jozo"
}`
	buf, err := out.Encode()
	assert.Nil(t, err)
	assert.Equal(t, expected, testu.PpBuf(buf))

	//testu.PPBuf(buf)
}

func TestFixtures(t *testing.T) {
	f := "kupasi"
	j := readFixture(t, f)

	e := NewExtender(j)
	out := e.ExtendWith(livescoreKyes)

	buf, err := out.Encode()
	assert.Nil(t, err)

	testu.AssertFixture(t,
		fmt.Sprintf("./fixtures/%s_out.json", f),
		[]byte(testu.PpBuf(buf)), saveFixtures)

	//testu.PPBuf(buf)
}

func readFixture(t *testing.T, fn string) *simplejson.Json {
	input, err := ioutil.ReadFile(fmt.Sprintf("./fixtures/%s.json", fn))
	assert.Nil(t, err)

	j, err := simplejson.NewJson(input)
	assert.Nil(t, err)
	return j
}

var livescoreKyes = map[string]string{
	"S": "sports",
	"C": "categories",
	"T": "tournaments",
	"M": "matches",
	"O": "scores",
	"A": "statistics",
	"G": "goals",
	"R": "cards",
	"U": "substitutions",
	"L": "lineups",
	"1": "team1",
	"2": "team2",
	"i": "id",
	"c": "city",
	"o": "country",
	"u": "currentPeriodStart",
	"d": "date",
	"n": "name",
	"s": "shotsOffGoal",
	"h": "shotsOnGoal",
	"t": "stadium",
	"a": "status",
	"y": "type",
	"v": "venue",
	"b": "ballPossesion",
	"w": "winner",
	"e": "team",
	"m": "time",
	"p": "player",
	"l": "playerIn",
	"r": "playerOut",
	"!": "playerTeam",
	"%": "pos",
	"'": "shirtNumber",
	"(": "substitute",
	"f": "referee",
	"k": "cornerKicks",
	")": "fouls",
	"*": "freeKicks",
	"g": "goalKicks",
	"+": "goalkeeperSaves",
	",": "offsides",
	"-": "throwIns",
	"0": "assist",
	"3": "lastGoal",
	"4": "from",
}
