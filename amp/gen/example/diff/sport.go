package diff

//go:generate go run gen.go

type Book struct {
	Version int
	Sports  Sports
	Events  Events `json:"events,omitempty"`
	Items   Items  `json:"events,omitempty"`
}

type (
	Sports     map[string]Sport
	Events     map[string]Event
	Categories map[string]Category
	Items      map[string]Item
)

type Sport struct {
	Name       string
	Categories Categories
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

type Item struct {
	Filed1 string
	Filed2 int
}
