* Nsq primjeri

1. Pokrenem consul, nsqd:
     ./start
2. U jednom treminalu:
     go run nsq_sub.go
   U drugom: 
     go run nsq_pub.go


* Leadership primjer

1. Pokrenem consul:
  ./start
2. U dva terminala dignem aplikaciju:
  go run leader_example.go
3. Zaustavim aplikaciju koja je leader, prezume drugi.
4. Pokrenem ponovo dvije. Nadjem pid-ove:
     ps aux | grep leader_example$
   Posaljem USR1 signal trenutnom leaderu npr.:
     kill -s USR1 70854
   Drugi preuzme leadership prvi ceka, posaljem signal drugom:
     kill -s USR1 70862
   I ponovim par puta jedno drugo.


* Http interface primjer

1. go run httpi
2. Navigiraj na:
   http://localhost:8123/ping         (curl -v pa pogledaj header Application)
   http://localhost:8123/health_check
   http://localhost:8123/debug/vars   (pogledaj svckit.stats key i kako je implementirano)
   http://localhost:8123/debug/pprof


* Metric

1. ./start
2. go run metric_example
3. Metrike su zapisane su poslane na udp 8125, tamo slusa statsd_to_nsq koji ih zapise u nsq kanal. Tail-am taj nsq kanal i tamo ih vidim:
     nsq_tail -topic=stats -lookupd-http-address=localhost:4161