package log2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/minus5/svckit/env"
	log "github.com/minus5/svckit/log"
	"github.com/stretchr/testify/assert"
)

/*
func TestSyslog(t *testing.T) {
	sysLog, err := syslog.Dial("udp", "10.0.66.192:514", syslog.LOG_LOCAL5, "testtag")
	fmt.Println(sysLog, err)
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
	}
	sysLog.Write([]byte("test slanja"))
	fmt.Println(sysLog, "This is a daemon warning with demotag.")
	//sysLog.Emerg("And this is a daemon emergency with demotag.")
}
*/

func TestCompare(t *testing.T) {
	n := 1

	//log.I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")
	//log.I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")

	var buf bytes.Buffer
	a := NewAgregator(&buf, 1)
	a.I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")
	outputlog2 := buf.String()

	outputlog := captureOutput(func() {
		log.I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")
	})

	logkv := parse(outputlog)
	log2kv := parse(outputlog2)

	if !assert.Equal(t, logkv["puta"], log2kv["puta"]) {
		fmt.Println("error u intu")
	}

	if !assert.Equal(t, logkv["pero"], log2kv["pero"]) {
		fmt.Println("error u s")
	}

	if !assert.Equal(t, logkv["float64"], log2kv["float64"]) {
		fmt.Println("error u float")
	}

	if !assert.Equal(t, logkv["host"], log2kv["host"]) {
		fmt.Println("error host")
	}

	if !assert.Equal(t, logkv["app"], log2kv["app"]) {
		fmt.Println("error app")
	}

	if !assert.Equal(t, logkv["msg"], log2kv["msg"]) {
		fmt.Println("error msg")
	}
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func parse(s string) map[string]interface{} {
	var data map[string]interface{}
	_ = json.Unmarshal([]byte(s), &data)
	return data
}

func TestCallerDepth(t *testing.T) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.CallerKey = "file"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, _ := config.Build(zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	))

	defer logger.Sync()

	logger.Info("level 0")

}

//obrisat println-ove!!!!!!!
func TestWrite(t *testing.T) {
	buf := `{"time":"2017-04-28T15:28:07.275599+02:00", "file":"requester_ss.go:269", "host":"app1", "app":"backend_api", "level":"info", "url":"http://upis_validator.service.sd/validacija", "listic_id":"132ABD96-883A-4781-A50E-CDF6029D326A", "request":"{\"listic_id\":\"132ABD96-883A-4781-A50E-CDF6029D326A\",\"listic\":{\"ulog_bez\":3.8,\"ulog_mt\":0.2,\"ulog_po_listicu\":0,\"tip\":0,\"tecaj\":162.3446,\"dobitak\":616.91,\"isplata\":555.6,\"porez\":61.31,\"tecajevi\":[592575030,592697949,592698007,592573743,592573706,592573821,592771883,592757192,592761762,592094372,592695621,592228284,592347751,592714900,592574718,592228182,592228104,592639122,592099161,592396633,592745295,592751139,592054510,592182036,592592143,592639065,592697544,592697561,592031487,592697111],\"fixevi\":null},\"listicTip\":\"sport\",\"tecajevi\":{\"592031487\":{\"tecaj_id\":592031487,\"ponuda_id\":10484743,\"dogadjaj_id\":10484743,\"tip\":\"1\",\"naziv\":\"B.Dortmund-1.FC K\u00f6ln\",\"broj\":\"625\",\"ponuda_db_id\":10484743,\"vrijeme\":\"29.04. 15:30\",\"tecaj\":1.25,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592054510\":{\"tecaj_id\":592054510,\"ponuda_id\":10408864,\"dogadjaj_id\":10408864,\"tip\":\"1\",\"naziv\":\"RB Salzburg-SV Ried\",\"broj\":\"322\",\"ponuda_db_id\":10408864,\"vrijeme\":\"29.04. 18:30\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592094372\":{\"tecaj_id\":592094372,\"ponuda_id\":10580557,\"dogadjaj_id\":10580557,\"tip\":\"1\",\"naziv\":\"BenficaL-GD Estoril Praia\",\"broj\":\"671\",\"ponuda_db_id\":10580557,\"vrijeme\":\"29.04. 19:15\",\"tecaj\":1.15,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592099161\":{\"tecaj_id\":592099161,\"ponuda_id\":10457768,\"dogadjaj_id\":10457768,\"tip\":\"1\",\"naziv\":\"Sheffield Utd-Chesterfield\",\"broj\":\"389\",\"ponuda_db_id\":10457768,\"vrijeme\":\"30.04. 13:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592182036\":{\"tecaj_id\":592182036,\"ponuda_id\":10448158,\"dogadjaj_id\":10448158,\"tip\":\"1\",\"naziv\":\"Pardub-1.SK Prost\u011bjov\",\"broj\":\"443\",\"ponuda_db_id\":10448158,\"vrijeme\":\"29.04. 10:15\",\"tecaj\":1.25,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592228104\":{\"tecaj_id\":592228104,\"ponuda_id\":10416699,\"dogadjaj_id\":10416699,\"tip\":\"1\",\"naziv\":\"HNK Rijeka-Inter Zapre\u0161i\u0107\",\"broj\":\"811\",\"ponuda_db_id\":10416699,\"vrijeme\":\"29.04. 19:00\",\"tecaj\":1.15,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592228182\":{\"tecaj_id\":592228182,\"ponuda_id\":10416703,\"dogadjaj_id\":10416703,\"tip\":\"1\",\"naziv\":\"Dinamo Zagreb-Cibalia\",\"broj\":\"815\",\"ponuda_db_id\":10416703,\"vrijeme\":\"29.04. 15:00\",\"tecaj\":1.1,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592228284\":{\"tecaj_id\":592228284,\"ponuda_id\":10421316,\"dogadjaj_id\":10421316,\"tip\":\"1\",\"naziv\":\"De Graafschap-Achilles\",\"broj\":\"618\",\"ponuda_db_id\":10421316,\"vrijeme\":\"28.04. 20:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592347751\":{\"tecaj_id\":592347751,\"ponuda_id\":10570435,\"dogadjaj_id\":10570435,\"tip\":\"1\",\"naziv\":\"B.Honv\u00e9d FC-Gyirm\u00f3t FC\",\"broj\":\"497\",\"ponuda_db_id\":10570435,\"vrijeme\":\"29.04. 18:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592396633\":{\"tecaj_id\":592396633,\"ponuda_id\":11601049,\"dogadjaj_id\":11601049,\"tip\":\"1\",\"naziv\":\"AhlyCairo-Entag El Harby\",\"broj\":\"770\",\"ponuda_db_id\":11601049,\"vrijeme\":\"30.04. 19:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592573706\":{\"tecaj_id\":592573706,\"ponuda_id\":10629183,\"dogadjaj_id\":10629183,\"tip\":\"2\",\"naziv\":\"Espanyol Bar.-FC Barcelona\",\"broj\":\"504\",\"ponuda_db_id\":10629183,\"vrijeme\":\"29.04. 20:45\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592573743\":{\"tecaj_id\":592573743,\"ponuda_id\":10629185,\"dogadjaj_id\":10629185,\"tip\":\"1\",\"naziv\":\"Real Madrid-CF Valencia\",\"broj\":\"506\",\"ponuda_db_id\":10629185,\"vrijeme\":\"29.04. 16:15\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592573821\":{\"tecaj_id\":592573821,\"ponuda_id\":10629186,\"dogadjaj_id\":10629186,\"tip\":\"1\",\"naziv\":\"Real Sociedad-Granada CF\",\"broj\":\"507\",\"ponuda_db_id\":10629186,\"vrijeme\":\"29.04. 13:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592574718\":{\"tecaj_id\":592574718,\"ponuda_id\":12041422,\"dogadjaj_id\":12041422,\"tip\":\"1\",\"naziv\":\"Kairat Almaty-FK Akzhayik\",\"broj\":\"657\",\"ponuda_db_id\":12041422,\"vrijeme\":\"29.04. 14:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592575030\":{\"tecaj_id\":592575030,\"ponuda_id\":12197388,\"dogadjaj_id\":12197388,\"tip\":\"1\",\"naziv\":\"ShakhtarD-FC Oleksandria\",\"broj\":\"824\",\"ponuda_db_id\":12197388,\"vrijeme\":\"30.04. 16:00\",\"tecaj\":1.15,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592592143\":{\"tecaj_id\":592592143,\"ponuda_id\":10351359,\"dogadjaj_id\":10351359,\"tip\":\"1\",\"naziv\":\"Monaco-Toulouse\",\"broj\":\"903\",\"ponuda_db_id\":10351359,\"vrijeme\":\"29.04. 17:00\",\"tecaj\":1.25,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592639065\":{\"tecaj_id\":592639065,\"ponuda_id\":10732352,\"dogadjaj_id\":10732352,\"tip\":\"1\",\"naziv\":\"Panathin.Atena-AE Larissa\",\"broj\":\"523\",\"ponuda_db_id\":10732352,\"vrijeme\":\"30.04. 18:00\",\"tecaj\":1.25,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592639122\":{\"tecaj_id\":592639122,\"ponuda_id\":10732358,\"dogadjaj_id\":10732358,\"tip\":\"1\",\"naziv\":\"PAOK Solun-AOK Kerkyra\",\"broj\":\"526\",\"ponuda_db_id\":10732358,\"vrijeme\":\"30.04. 18:00\",\"tecaj\":1.1,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592695621\":{\"tecaj_id\":592695621,\"ponuda_id\":10769457,\"dogadjaj_id\":10769457,\"tip\":\"1\",\"naziv\":\"Wolfsburg II-SV Eichede\",\"broj\":\"1819\",\"ponuda_db_id\":10769457,\"vrijeme\":\"29.04. 13:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592697111\":{\"tecaj_id\":592697111,\"ponuda_id\":11595564,\"dogadjaj_id\":11595564,\"tip\":\"1\",\"naziv\":\"Al Ahli Dubai-Al Sharjah\",\"broj\":\"1192\",\"ponuda_db_id\":11595564,\"vrijeme\":\"29.04. 16:00\",\"tecaj\":1.25,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592697544\":{\"tecaj_id\":592697544,\"ponuda_id\":11653999,\"dogadjaj_id\":11653999,\"tip\":\"1\",\"naziv\":\"LevadTal-P\u00e4rnu JK Vaprus\",\"broj\":\"779\",\"ponuda_db_id\":11653999,\"vrijeme\":\"28.04. 18:00\",\"tecaj\":1.03,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592697561\":{\"tecaj_id\":592697561,\"ponuda_id\":11654001,\"dogadjaj_id\":11654001,\"tip\":\"2\",\"naziv\":\"ViljanTul-Flora Tallinn\",\"broj\":\"780\",\"ponuda_db_id\":11654001,\"vrijeme\":\"28.04. 18:00\",\"tecaj\":1.15,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592697949\":{\"tecaj_id\":592697949,\"ponuda_id\":11983617,\"dogadjaj_id\":11983617,\"tip\":\"1\",\"naziv\":\"Buriram Utd-Sisaket FC\",\"broj\":\"848\",\"ponuda_db_id\":11983617,\"vrijeme\":\"29.04. 15:00\",\"tecaj\":1.1,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592698007\":{\"tecaj_id\":592698007,\"ponuda_id\":11983620,\"dogadjaj_id\":11983620,\"tip\":\"1\",\"naziv\":\"MuangTU-Suphanburi FC\",\"broj\":\"851\",\"ponuda_db_id\":11983620,\"vrijeme\":\"30.04. 14:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592714900\":{\"tecaj_id\":592714900,\"ponuda_id\":12375577,\"dogadjaj_id\":12375577,\"tip\":\"1\",\"naziv\":\"Keflav\u00edk-Vidir Gardur\",\"broj\":\"254\",\"ponuda_db_id\":12375577,\"vrijeme\":\"28.04. 21:00\",\"tecaj\":1.2,\"dogadjaj_grupa\":5,\"sport_razrada_id\":22},\"592745295\":{\"tecaj_id\":592745295,\"ponuda_id\":10448153,\"dogadjaj_id\":10448153,\"tip\":\"1\",\"naziv\":\"Olomouc-Ban\u0...", "msg":"validate_service"}`

	len, err := out.Write([]byte(buf))
	assert.Equal(t, len, 7768)
	assert.Equal(t, err, nil)
}

func TestSplitLevelMessage(t *testing.T) {
	data := []struct {
		line  string
		level string
		msg   string
	}{
		{"[DEBUG] nesto", LevelDebug, "nesto"},
		{"[NOTICE] nesto", LevelNotice, "nesto"},
		{"[ERROR] nesto", LevelError, "nesto"},
		{"error nesto", LevelError, "error nesto"},
		{"pero nesto", LevelDebug, "pero nesto"},
	}

	for _, d := range data {
		level, msg := splitLevelMessage(d.line)
		assert.Equal(t, d.level, level)
		assert.Equal(t, d.msg, msg)
	}
}

func TestInfoLog0(t *testing.T) {
	S("pero", "zdero").Info("prvi")
	S("pero", "zdero").Info("drugi")
	S("pero", "zdero").Info("treci")
}

func TestInfoLog(t *testing.T) {
	logger := func() *Agregator {
		return S("pero", "zdero")
	}
	logger().Info("prvi")
	logger().Info("drugi")
	logger().Info("treci")
}

func TestInfoLog2(t *testing.T) {
	logger := S("pero", "zdero") //.New()
	logger.Info("prvi")
	logger.Info("drugi")
	logger.Info("treci")

	logger.Info("prvi")
	logger.Info("drugi")
	logger.Info("treci")

	logger.Info("prvi")
	logger.Info("drugi")
	logger.Info("treci")
}

func BenchmarkSvckitLog(b *testing.B) {
	for n := 0; n < b.N; n++ {
		I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")
	}
}

func BenchmarkZap2(b *testing.B) {
	//startProfile()
	//for n := 0; n < b.N; n++ {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	//cfg.EncoderConfig.LevelKey = "level"
	//cfg.EncoderConfig.MessageKey = "msg"
	//cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//cfg.Development = false
	//cfg.Encoding = "json"
	logger, _ := cfg.Build(zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	))
	defer logger.Sync()
	//sugar := logger.Sugar()

	for n := 0; n < b.N; n++ {
		logger.Info("iso medo u ducan",
			zap.Int("puta", n),
			zap.Float64("float64", 3.1415926535),
			zap.String("pero", "zdero"),
			zap.String("key", "value"),
		)
	}
	//stopProfile()
}

func BenchmarkZapSvckit(b *testing.B) {
	//startProfile()
	//a := newAgregator(2)
	//fmt.Println(a)
	for n := 0; n < b.N; n++ {
		//Info("msg")
		I("puta", n).S("pero", "zdero").F("float64", 3.1415926535, -1).S("key", "value").Info("iso medo u ducan")
	}
	//stopProfile()
}

func TestZap(t *testing.T) {
	//startProfile()
	//n := 1
	//I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Notice("iso medo u ducan")
	Printf("hej")
	Info("msg")
	//stopProfile()
	//I("puta", 1).Debug("msg")
	F("float64", 3.1415926535, -1).Info("msg")
	S("pero", "zdero").Info("msg")
	S("key", "value").Notice("iso medo u ducan")
	//Debug("msg")
	Info("msg")
	Notice("msg")
	//Errorf("msg")
}

func TestError(t *testing.T) {
	Errorf("msg")
	Error(nil)

}

func BenchmarkKeyValue4(b *testing.B) {
	a := Agregator{}
	for n := 0; n < b.N; n++ {
		a.fields = append(a.fields, zap.Int("key", 5))
		//a.fields = append(a.fields, zap.Int("key", 5))
		//a.fields = append(a.fields, zap.Int("key", 5))
		//a.fields = append(a.fields, zap.Int("key", 5))
		a.fields = nil
	}
}

/*
func startProfile() {
	//output := fmt.Sprintf("/Users/antonio/work/pprof/log/%s.pprof", time.Now().Format(time.RFC3339))
	output := fmt.Sprintf("/Users/antonio/work/pprof/log/a.pprof")
	log.S("output", output).Info("starting profile")
	// msgs := reply.New(profileDir)
	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal(err)
	}
}

func stopProfile() {
	pprof.StopCPUProfile()
}
*/
func TestUse(t *testing.T) {
	// globalni
	S("pero", "zdero").Info("1")
	S("jozo", "bozo").S("jozo1", "bozo1").Info("2")

	// lokalni
	logger := New()                     // kreira novu instancu agregatora i zap-a
	logger.S("pero", "zdero").Info("1") // nova instanca agregagtora
	logger.S("jozo", "bozo").Info("2")

	// lokalni sa zajednickim atributima
	logger = S("part", "1").New()
	logger.S("pero", "zdero").S("1", "1").Info("1")
	logger.S("jozo", "bozo").S("2", "2").Info("2")

	func() {
		logger.S("bozo", "misteriozo").Info("3")
	}()
}

func TestLocal(t *testing.T) {
	logger := S("part", "1").New()
	logger.S("pero", "zdero").S("1", "1").Info("1")
	logger.S("jozo", "bozo").S("2", "2").Info("2")
}

func TestPrint(t *testing.T) {
	Printf("[INFO] pero zdero")
	Printf("[NOTICE] pero zdero %d", 123)
	Printf("[NOTICE] sto bude kada u istoj app koristim classic logger")
}

func TestLogLog2routines(t *testing.T) {
	var wg sync.WaitGroup

	for j := 0; j < 5; j++ {
		wg.Add(1)
		go l2(strconv.Itoa(j))
		go l(strconv.Itoa(j))
		wg.Done()
	}
	wg.Wait()
}

func l2(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		S("log2 rutina", s).I("poziv", i).Info("msg")
	}
}

func l(s string) {
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		log.S("log rutina", s).I("poziv", i).Info("msg")
	}
}
