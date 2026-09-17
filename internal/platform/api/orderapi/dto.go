package orderapi

type CreateOrderRequest struct {
	RestaurantID string          `json:"restaurantId"`
	Items        []OrderItemIn   `json:"items"`
}

type OrderItemIn struct {
	ItemID string `json:"itemId"`
	Qty    int    `json:"qty"`
}

type Order struct {
	ID             string      `json:"id"`
	CustomerID     string      `json:"customerId"`
	RestaurantID   string      `json:"restaurantId"`
	RestaurantName string      `json:"restaurantName"`
	Status         string      `json:"status"`
	TotalMinor     int64       `json:"totalMinor"`
	CourierID      string      `json:"courierId,omitempty"`
	PaymentID      string      `json:"paymentId,omitempty"`
	Items          []OrderItem `json:"items"`
	CreatedAt      string      `json:"createdAt"`
	UpdatedAt      string      `json:"updatedAt"`
}

type OrderItem struct {
	ItemID         string `json:"itemId"`
	Name           string `json:"name"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
	Qty            int    `json:"qty"`
}

type OrderList struct {
	Items []Order `json:"items"`
}
