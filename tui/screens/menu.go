package screens

import "github.com/CheeziCrew/curd"

// MenuModel wraps curd.MenuModel for gruyere.
type MenuModel = curd.MenuModel

// NewMenu creates a fresh menu model.
func NewMenu() curd.MenuModel {
	return curd.NewMenuModel(curd.MenuConfig{
		Banner: []string{
			"                                        ",
			"   __ _ _ __ _   _ _   _  ___ _ __ ___  ",
			"  / _` | '__| | | | | | |/ _ \\ '__/ _ \\ ",
			" | (_| | |  | |_| | |_| |  __/ | |  __/ ",
			"  \\__, |_|   \\__,_|\\__, |\\___|_|  \\___| ",
			"  |___/            |___/                  ",
		},
		Tagline: "age your APIs gracefully",
		Palette: curd.GruyerePalette,
		Items: []curd.MenuItem{
			{Icon: "🧀", Name: "Generate Changelog", Command: "changelog", Desc: "diff OpenAPI specs between branches"},
		},
	})
}
