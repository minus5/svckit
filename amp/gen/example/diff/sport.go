package diff

//go:generate go run gen.go

type Book struct {
	Version int
	Sports  map[string]Sport
	Events  map[string]Event `json:"events,omitempty"`
}

type Sport struct {
	Name       string
	Categories map[string]Category
	Order      int `json:",omitempty"`
}

type Category struct {
	Name  string
	Order int `json:",omitempty"`
}

type Event struct {
	Home   string `json:",omitempty"`
	Away   string `json:",omitempty"`
	Result Result
}

type Result struct {
	Home int
	Away int
}
