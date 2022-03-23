# svckit examples

## NSQ example

Start consul:
```
./start
```

Note: consul agent can be also started in development mode using command:
```
consul agent -data-dir=tmp/consul -dev
```
In first terminal run:
```
go run nsq_sub.go
```     
In second terminal run:
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
Start two apps again, find the PIDs:
```
ps aux | grep leader_example$
```     
Send USR1 signal to current leader, e.g.:
```   
kill -s USR1 70854
```  
Second takes over the lead, first one is waiting. Send the signal to current leader,e .g.:
```   
kill -s USR1 70862
```  
Roles will switch again: First takes over the lead back, second one is waiting.


## Http interface example
```
go run httpi_example.go
```
Then run:
```
http://localhost:8123/ping
```
Notice the header *Application* in the response. Also try running following commands:

```
http://localhost:8123/health_check
```

```
 http://localhost:8123/debug/vars
```

```
http://localhost:8123/debug/pprof
```

## Metrics example

Start consul:
```
 ./start
```
The run:
```
go run metric_example.go
``` 
The metrics are written and sent to *udp:8125* port. Here *statsd_to_nsq* is listening which writes them into NSQ channel. To see more run:
```
nsq_tail -topic=stats -lookupd-http-address=localhost:4161
```