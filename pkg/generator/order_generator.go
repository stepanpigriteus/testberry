package generator

import (
	"math/rand"
	"strconv"
	order_entity "testberry/internal/domain/order"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func GenerateRandomOrder(seed int64) order_entity.Order {
	source := rand.NewSource(seed)
	r := rand.New(source)

	orderUID := RandString(r, 20)

	return order_entity.Order{
		OrderUID:    orderUID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: order_entity.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: order_entity.Payment{
			Transaction:  orderUID,
			RequestID:    "REQ" + strconv.Itoa(rand.Intn(10000000)),
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []order_entity.Item{
			{
				ChrtID:      rand.Intn(10000000),
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         RandString(r, 20),
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}
