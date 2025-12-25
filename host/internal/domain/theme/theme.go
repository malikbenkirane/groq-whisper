package theme

type Description struct {
	Name       Name
	Title      string
	Categories []Category
}

type Category struct {
	Name     string
	Keywords []Keyword
}

type Keyword string
type Name string
