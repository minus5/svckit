### Zasto
  * da ne ponavljamo isti kod u svakom servisu
  * da ga probamo izvuci na zajednicko mjesto

### Ciljevi
  * lako je proizvesti novi servis 
  * jednostavno za koristenje
  * sto manje koda za novi servis

### Funkcionalnosti
  * loging
  * service discovery
  * nsq
  * metric
  * pprof
  * health_check
  * ping
  * expvar
  * leader election


### Environment varijable

Environment varijable koje utjecu na ponsanje:

   * $dc      - naziv datacentra u kome radi aplikacija
   * $node    - ime hosta na kome se nalazi docker
   * $node_ip - vanjska ip adresa node-a
   * $sd_dns  - adresa na kojoj se nalazi service discovery (consul) dns

Node i node_ip varijable imaju smisla kada se aplikacija vrti unutar docker containera. Tako saznajem na kojem node-u se aplikacija vrti. Node je ovdje naziv za stroj na kome se nalazi docker. Cesto mi je potrebna vanjska ip adresa. Unutar containera mogu saznati samo lokalnu, pa ju na ovaj nacin prosljedjujem.


### Preduvjeti

Da bi svckit radio bitno mi je pronaci consul kojeg cu pitati za druge servise. Npr. pitat cu ga za lokaciju nsqlookup-a.

Ocekujem da je u consulu registriran barem jedan servis *nsqlookupd-http.service.sd* inace ne fukcionira nsq.

Bitno je gdje mi se nalazi nsqd na koji cu pisati. Defaultno lokalno ili na $node_ip ako je definirana ta varijabla.

Syslog na koji pisem je localhost:514 ili $node_ip:514.

