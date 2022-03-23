# svckit examples

## NSQ example

Start consul, NSQ daemon:
```
./start
```
In first terminal
```
go run nsq_sub.go
```     
In second terminal:
```
go run nsq_pub.go
```


## Leadership example

Start consul:
```
  ./start
```  
Run the application in two terminals:
```
  go run leader_example.go
```  
Stop the application which becomes the leader. The other one will take over.
Start two apps again, find the pids:
```
ps aux | grep leader_example$
```     
Send USR1 signal to current leader, e.g.:
```   
kill -s USR1 70854
```  
Second takes over the lead, first one is waiting. Send the signal to other,e .g.:
```   
kill -s USR1 70862
```  
And repeat couple times.


## Http interface example
```
go run httpi
```
Go to:

* http://localhost:8123/ping         (curl -v see header Application)
* http://localhost:8123/health_check
* http://localhost:8123/debug/vars   (See svckit.stats key i how it is implemented)
* http://localhost:8123/debug/pprof


## Metrics example

Start consul:
```
 ./start
```
The run:
```
go run metric_example
``` 
The metrics are written and sent to udp 8125. Here *statsd_to_nsq* is listening which writes the metrics to NSQ channel. To see more run:
```
nsq_tail -topic=stats -lookupd-http-address=localhost:4161
```