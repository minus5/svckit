package diff

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testFulls = []string{
	`{
		"Version": 1,
		"Sports": {
			"1": {
				"Name": "soccer",
				"Categories": {
					"1": {
						"Name": "England"
					}
				}
			}
		}
	}`, `{
  "Version": 2,
  "Sports": {
    "1": {
      "Name": "soccer",
      "Categories": {
        "1": {
          "Name": "Croatia"
        }
      }
    }
  }
}`, `{
  "Version": 2,
  "Sports": {
    "1": {
      "Name": "soccer",
      "Categories": {
        "1": {
          "Name": "Croatia"
        },
        "2": {
          "Name": "Spain"
        }
      }
    }
  }
}`, `{
  "Version": 2,
  "Sports": {
    "1": {
      "Name": "soccer",
      "Categories": {
        "2": {
          "Name": "Spain"
        }
      }
    }
  }
}`,
}

var testDiffs = []string{`{
  "version": 2,
  "sports": {
    "1": {
      "categories": {
        "1": {
          "name": "Croatia"
        }
      }
    }
  }
}`, `{
  "version": 2,
  "sports": {
    "1": {
      "categories": {
        "1": {
          "name": "Croatia"
        },
        "2": {
          "name": "Spain"
        }
      }
    }
  }
}`, `{
  "version": 2,
  "sports": {
    "1": {
      "categories": {
        "1": null,
        "2": {
          "name": "Spain"
        }
      }
    }
  }
}`,
}

func book() Book {
	b := Book{}
	err := json.Unmarshal([]byte(testFulls[0]), &b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func diff(no int) *BookDiff {
	b := &BookDiff{}
	err := json.Unmarshal([]byte(testDiffs[no]), b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func TestCreateDiff(t *testing.T) {
	b1 := book()
	b2 := book()

	b2.Version = 2
	b2.Sports["1"].Categories["1"] = Category{Name: "Croatia"}

	i := b1.Diff(b2)
	assert.Equal(t, testDiffs[0], js(i))

	b2.Sports["1"].Categories["2"] = Category{Name: "Spain"}
	i = b1.Diff(b2)
	assert.Equal(t, testDiffs[1], js(i))

	delete(b2.Sports["1"].Categories, "1")
	i = b1.Diff(b2)
	assert.Equal(t, testDiffs[2], js(i))
	//	pp(i)
}

func TestMerge(t *testing.T) {
	b := book()
	i := diff(0)
	b, _ = b.MergeDiff(i)
	assert.Equal(t, testFulls[1], js(b))

	b = book()
	i = diff(1)
	b, _ = b.MergeDiff(i)
	assert.Equal(t, testFulls[2], js(b))

	b = book()
	i = diff(2)
	b, _ = b.MergeDiff(i)
	assert.Equal(t, testFulls[3], js(b))
	//pp(b)

	var b1 Book
	b2, _ := b1.MergeDiff(diff(1))
	assert.Len(t, b1.Sports, 0)
	assert.Len(t, b2.Sports, 1)
	// pp(b1)
	// pp(b2)
}

func pp(o interface{}) {
	fmt.Printf("%s\n", js(o))
}

func js(o interface{}) string {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(buf)
}
