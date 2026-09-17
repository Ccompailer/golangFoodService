package catalogapi

type Restaurant struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Cuisine     string     `json:"cuisine"`
	Address     string     `json:"address"`
	IsOpen      bool       `json:"isOpen"`
	MinOrder    int64      `json:"minOrderMinor"`
	PrepMinutes int        `json:"prepMinutes"`
	Menu        []Category `json:"menu,omitempty"`
}

type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

type Item struct {
	ID          string   `json:"id"`
	RestaurantID string  `json:"restaurantId,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceMinor  int64    `json:"priceMinor"`
	Allergens   []string `json:"allergens"`
	Available   bool     `json:"available"`
}

type RestaurantList struct {
	Items []Restaurant `json:"items"`
}

type ItemList struct {
	Items []Item `json:"items"`
}
