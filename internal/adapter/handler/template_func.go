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
		{Value: "coffee", Label: "Coffee"},
		{Value: "emerald", Label: "Emerald"},
		{Value: "corporate", Label: "Corporate"},
		{Value: "lofi", Label: "Lofi"},
		{Value: "business", Label: "Business"},
		{Value: "winter", Label: "Winter"},
		{Value: "nord", Label: "Nord"},
		{Value: "aqua", Label: "Aqua"},
		{Value: "luxury", Label: "Luxury"},
		{Value: "dim", Label: "Dim"},
		{Value: "night", Label: "Night"},
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
