# svckit examples

Before running individual examples run in separate terminal following command:
```
./start
```

This will internally run [goreman](https://github.com/mattn/goreman) and launch all necessary services (NSQ services + consul) as defined in [Procfile](./Procfile). Consul agent will be started with services defined in [consul.json](./consul.json). 

Following admin UIs will be then running on localhost:

* **Consul UI**: http://localhost:8500/
* **NSQ Admin**: http://localhost:4171/

## NSQ example

In first terminal run:
```
go run nsq_sub.go
```     
In second terminal run:
```
go run nsq_pub.go
```


## Leadership example
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
Notice HTTP header *Application* in the response. Also try running following commands:

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
Run:
```
go run metric_example.go
``` 
The metrics are written and sent to *udp:8125* port. Here *statsd_to_nsq* is listening which writes them into NSQ channel. To see more run:
```
nsq_tail -topic=stats -lookupd-http-address=localhost:4161
```