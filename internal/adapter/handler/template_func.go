package handler

type ThemeOption struct {
	Value string
	Label string
}

func ThemeOptions() []ThemeOption {
	return []ThemeOption{
		{Value: "solvi", Label: "Solvi"},
		{Value: "wine", Label: "Wine"},
		{Value: "light", Label: "Light"},
		{Value: "dark", Label: "Dark"},
		{Value: "cupcake", Label: "Cupcake"},
		{Value: "bumblebee", Label: "Bumblebee"},
		{Value: "emerald", Label: "Emerald"},
		{Value: "corporate", Label: "Corporate"},
		{Value: "synthwave", Label: "Synthwave"},
		{Value: "retro", Label: "Retro"},
		{Value: "cyberpunk", Label: "Cyberpunk"},
		{Value: "valentine", Label: "Valentine"},
		{Value: "halloween", Label: "Halloween"},
		{Value: "garden", Label: "Garden"},
		{Value: "forest", Label: "Forest"},
		{Value: "aqua", Label: "Aqua"},
		{Value: "lofi", Label: "Lofi"},
		{Value: "pastel", Label: "Pastel"},
		{Value: "fantasy", Label: "Fantasy"},
		{Value: "wireframe", Label: "Wireframe"},
		{Value: "black", Label: "Black"},
		{Value: "luxury", Label: "Luxury"},
		{Value: "dracula", Label: "Dracula"},
		{Value: "cmyk", Label: "Cmyk"},
		{Value: "autumn", Label: "Autumn"},
		{Value: "business", Label: "Business"},
		{Value: "acid", Label: "Acid"},
		{Value: "lemonade", Label: "Lemonade"},
		{Value: "night", Label: "Night"},
		{Value: "coffee", Label: "Coffee"},
		{Value: "winter", Label: "Winter"},
		{Value: "dim", Label: "Dim"},
		{Value: "nord", Label: "Nord"},
		{Value: "sunset", Label: "Sunset"},
		{Value: "caramellatte", Label: "Caramellatte"},
		{Value: "abyss", Label: "Abyss"},
		{Value: "silk", Label: "Silk"},
	}
}

func MakeMapFunc(values ...interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue
		}
		m[key] = values[i+1]
	}
	return m, nil
}
