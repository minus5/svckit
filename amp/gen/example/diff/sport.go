package diff

//go:generate go run gen.go

type Book struct {
	Version int
	Sports  map[string]Sport
}

type Sport struct {
	Name       string
	Categories map[string]Category
}

type Category struct {
	Name string
}
