package theme

type Description struct {
	Name       string
	Title      string
	Categories []Category
}

type Category struct {
	Name     string
	Keywords []Keyword
}

type Keyword string
