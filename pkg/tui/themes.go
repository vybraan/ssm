package tui

type theme struct {
	titleColor     string
	selTitleColor  string
	selBorderColor string
	selDescColor   string
}

var themes = map[string]theme{
	"matrix": matrixTheme(),
	"sky":    skyTheme(),
}

func matrixTheme() theme {
	return theme{
		titleColor:     "#648c11",
		selTitleColor:  "#9efd38",
		selBorderColor: "#9efd38",
		selDescColor:   "#648c11",
	}
}

func skyTheme() theme {
	return theme{
		titleColor:     "#4682b4",
		selTitleColor:  "#00bfff",
		selBorderColor: "#00bfff",
		selDescColor:   "#4682b4",
	}
}
