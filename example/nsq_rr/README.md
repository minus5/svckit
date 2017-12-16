#### Pokretanje

1. Terminal
Pokrenem potrebne servise; consul, nsqd , nsqlookup tako da opalim skriptu start iz vendor/github.com/minus5/svckit/example.
```
cd ..  
./start
```

2. Terminal
Server i monitoring nsq kanala:
```
./start
```

3. Terminal
```
cd client
go run main.go
```

Prvo pokretanje ce stvoriti kanale, pa je ok pricekati 10-tak sekundi za odgovor.

#### Ideje

* Paziti na dependency:
	* api paket je neovisan o tehnologiji
	* client (client/main.go) i service (u server/main.go) barataju samo intristic tipovima (string, int)
	* sve vezano za nsq je u api/nsq
	* main-ovi client i server moraju biti jako jednostavni
	* novi trasport (http, ...) se moze dodati bez promjene api paketa, a client i server se promjena svodi na zamjenu paketa kojeg referenciraju
	
* Imati typed error-e. Iste i na  client i na server. Error se serijalizira (u string) i na drugoj strani raspakuje u typed. Tako da mogu u kodu negdje pitati if err == ErrNeki. Ovdje imamo primjer za ErrOverflow.

* Tipove atributa u porukama postaviti po tipu strukture.

* Konvencijom dodjeliti nazive response kanala.

