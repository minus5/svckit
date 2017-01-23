package v8w

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestJsSource = `27 + 15`
	TestExpect   = `42`
)

var v *V8W

func TestNew(t *testing.T) {
	var err error
	v, err = New(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	fmt.Printf("V8 version: %s\n", V8Version())
}

func TestEval(t *testing.T) {
	r, err := v.Eval(TestJsSource)
	assert.Nil(t, err)
	assert.Equal(t, TestExpect, r)
}

func TestEvalConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			fmt.Println("start", i)
			defer fmt.Println("done", i)
			r, err := v.Eval(TestJsSource)
			assert.Nil(t, err)
			assert.Equal(t, TestExpect, r)
			wg.Done()
		}(i)
	}
	wg.Add(1)
	go func() {
		fmt.Println("first start")
		defer fmt.Println("first end")
		r, err := v.Eval(TestJsSource)
		assert.Nil(t, err)
		assert.Equal(t, TestExpect, r)
		wg.Done()
	}()
	func() {
		fmt.Println("second start")
		defer fmt.Println("second end")
		r, err := v.Eval(TestJsSource)
		assert.Nil(t, err)
		assert.Equal(t, TestExpect, r)
	}()
	wg.Wait()
}

func TestListic(t *testing.T) {
	src, err := ioutil.ReadFile("/Users/ianic/work/services/src/cmd/backend_api/assets/listic_provjera.js")
	assert.Nil(t, err)
	o, err := New([]byte(string(src) + `;var worker = new require("Listic.lib/main")();`))

	assert.Nil(t, err)
	l := `worker.provjeri(
        {
			"ulog_bez":9.5,
			"ulog_mt":0.5,
			"ulog_po_listicu":0,
			"tip":0,
			"tecajevi":[533023808,533023844,533023862,533023867,533023874,533023831],
			"fixevi":null,
			"kombinacije":null
		}
,
{
			"533023808": {"tecaj_id":533023808,"ponuda_id":9931468,"dogadjaj_id":9931468,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":1.5,"max_utakmica":100},
			"533023844": {"tecaj_id":533023844,"ponuda_id":9931476,"dogadjaj_id":9931476,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":1.6,"max_utakmica":100},
			"533023862": {"tecaj_id":533023862,"ponuda_id":9931483,"dogadjaj_id":9931483,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":2.25,"max_utakmica":100},
			"533023867": {"tecaj_id":533023867,"ponuda_id":9931485,"dogadjaj_id":9931485,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"1","tecaj":1.6,"max_utakmica":100},
			"533023874": {"tecaj_id":533023874,"ponuda_id":9931488,"dogadjaj_id":9931488,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":2.05,"max_utakmica":100},
			"533023831": {"tecaj_id":533023831,"ponuda_id":9931472,"dogadjaj_id":9931472,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"1","tecaj":1.4,"max_utakmica":100}
		}
    );`
	rsp, err := o.Eval(l)
	assert.Nil(t, err)
	//fmt.Printf("%s\n", rsp)
	eRsp := `{"ok":true,"error":"","listic":{"tip":0,"ulog_bez":9.5,"ulog_mt":0.5,"tecaj":24.7968,"live":false,"tecajevi":[533023808,533023844,533023862,533023867,533023874,533023831],"dobitak":235.57,"isplata":212.96,"porez":22.61}}`

	assert.Equal(t, eRsp, rsp)

}

func BenchmarkListic(b *testing.B) {
	src, _ := ioutil.ReadFile("/Users/ianic/work/services/src/cmd/backend_api/assets/listic_provjera.js")
	o, _ := New(src)

	l := `
var worker = new require("Listic.lib/main")();
worker.provjeri(
        {
			"ulog_bez":9.5,
			"ulog_mt":0.5,
			"ulog_po_listicu":0,
			"tip":0,
			"tecajevi":[533023808,533023844,533023862,533023867,533023874,533023831],
			"fixevi":null,
			"kombinacije":null
		}
,
{
			"533023808": {"tecaj_id":533023808,"ponuda_id":9931468,"dogadjaj_id":9931468,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":1.5,"max_utakmica":100},
			"533023844": {"tecaj_id":533023844,"ponuda_id":9931476,"dogadjaj_id":9931476,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":1.6,"max_utakmica":100},
			"533023862": {"tecaj_id":533023862,"ponuda_id":9931483,"dogadjaj_id":9931483,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":2.25,"max_utakmica":100},
			"533023867": {"tecaj_id":533023867,"ponuda_id":9931485,"dogadjaj_id":9931485,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"1","tecaj":1.6,"max_utakmica":100},
			"533023874": {"tecaj_id":533023874,"ponuda_id":9931488,"dogadjaj_id":9931488,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"2","tecaj":2.05,"max_utakmica":100},
			"533023831": {"tecaj_id":533023831,"ponuda_id":9931472,"dogadjaj_id":9931472,"razrada_id":188186,"liga_id":188186,"sport_id":234,"tip":"1","tecaj":1.4,"max_utakmica":100}
		}
    );`

	for n := 0; n < b.N; n++ {
		_, _ = o.Eval(l)
		//eRsp := `{"error":"","listic":{"dobitak":235.57,"isplata":212.96,"live":false,"porez":22.61,"tecaj":24.7968,"tecajevi":[533023808,533023844,533023862,533023867,533023874,533023831],"tip":0,"ulog_bez":9.5,"ulog_mt":0.5},"ok":true}`
		//assert.Equal(t, eRsp, rsp)
	}
}
