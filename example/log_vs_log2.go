// +build test

package main

import (
	"log/syslog"
	"os"

	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/log2"
)

func main() {
	sysLog, _ := syslog.Dial("udp", "10.0.66.192:514", syslog.LOG_LOCAL5, "test")

	//ispis u syslog koristenjem log2 paketa upotrebom rucno dodanog syslog writera
	l2 := log2.NewAgregator(sysLog, 1)
	l2.Info("msg")

	//ispis u syslog koristenjem log paketa upotrebom rucno dodanog syslog writera
	l := log.NewAgregator(sysLog, 3)
	l.Info("msg")

	//klasicni poziv log2 paketa koji ispisuje na unaprijed definirani izlaz
	l2.I("puta", 1).F("float64", 3.1415926535, 1).S("key", "val").Info("msg3")

	//klasicni poziv log paketa koji ispisuje na unaprijed definirani izlaz
	l.I("puta", 1).F("float64", 3.1415926535, 1).S("key", "val").Info("msg4")

	//ispis u stderr koristenjem log2 paketa gdje je writer rucno postavljen na stderr
	l2 = log2.NewAgregator(os.Stderr, 1)
	l2.I("puta", 1).F("float64", 3.1415926535, 1).S("key", "val").Info("msg5")

	//ispis u stderr pomocu postojeceg log paketa gdje je writer postavljen na stderr
	l = log.NewAgregator(os.Stderr, 3)
	l.I("puta", 1).F("float64", 3.1415926535, 1).S("key", "val").Info("msg6")

}
