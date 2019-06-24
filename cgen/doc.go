/*
Package cgen generira kod za rad s immutable strukturama.

U Go-u value objekti su imutable. Problemcic je ako sadrze map ili slice. Oni su
reference types. Kopiranje napravi novu instancu mape ali onda i dalje pokazuje
na isti komad memorije. Pa nije kopirana. Modifikacija u kopiji radi
modifikaciju u originalu.

Kod immutable uvijek prizvodimo novu kopiju. Nikada ne modificiramo postojeci
objekt. Tako da nema potrebe za sinkronizacijom medju konkurentnim goroutinama.
Izbjegavamo sve mutexe, lockove, serijalizaciju citanja i pisanja. Jesi li ikad
u produkciji dobio: "panic: concurrent map read and map write", uz malo pravila
mozemo napraviti uvjete u kojima takav panic nije moguce proizvesti.

Ovaj paket daje dva nacina da dobijemo stvarno immutable objekte.

1. deep copy

Genrator napravi Copy metodu na svakom objetku ona napravi kopiju svih inner
mapa i dalje rekurzivno. Takva kopija sigurna je za konkuretno koristenje.

2. merge diff-a

Iz inicijalne strukture generiramo diff strukturu. U instanci diff strukture
opisemo razlike izmedju trenutnog i zeljenog stanja. Merge tog diff-a na
inicijalnoj strukturi proizvodi kopiju s izmjenjenim dijelovim. Dijelovi koji
nisu promjenjeni ostaju zajednicki i staroj i novoj instanci. Pogodno za velike
strukture koje imaju puno sitnih modifikacija.

Generator se ogranicava na odredjeni tip inicijalne strukture:
 - strukturu uvijek koristimo kao value, nikad kao pointer
 - svi fileds u struturi su value types
 - moze sadrzavati inner strukture koje su opet value type
 - moze sadrzavati mape koje takodjer imaju value type objekte za value
 - NE moze sadrzavati slice, array

Primjer podrzanog tipa strukture:

type Event struct {
	ID       int             // intrinstic value types
	Home     string
	Away     string
	Schedule time.Time
	Active   bool
	Result   Result          // inner struct of value types
	Markets  map[int]Market  // inner map of value types
}
type Market struct {
	Name     string
	Outcomes map[int]Outcome
}
type Outcome struct {
	Name string
	Odds float64
}
type Result struct {
	Home int
	Away int
}
*/
package cgen
