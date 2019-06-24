package example

//go:generate go run gen.go
// //go:generate easyjson -all event.go
// //go:generate easyjson -all event_diff_gen.go

type Event struct {
	Home    string
	Away    string
	Markets map[int]Market
}
type Market struct {
	Name     string
	Outcomes map[int]Outcome
}
type Outcome struct {
	Name string
	Odds float64
}
