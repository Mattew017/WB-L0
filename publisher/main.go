package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
	"web_service/internal/config"
	"web_service/internal/domain"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.Load()
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	order := generateRandomOrder()

	// Конвертируем в JSON
	jsonData, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal order: %v", err)
	}

	log.Println("Generated order:")
	log.Println(string(jsonData))

	topic := cfg.KafkaTopic
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: jsonData,
	}

	deliveryChan := make(chan kafka.Event)
	err = producer.Produce(message, deliveryChan)
	if err != nil {
		log.Fatalf("Failed to produce message: %v", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		log.Printf("Message delivered to %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)
}

func generateRandomOrder() domain.Order {
	orderUID := uuid.New().String()
	trackNumber := fmt.Sprintf("WBILM%s", uuid.New().String()[:8])
	now := time.Now()
	numItems := rand.Intn(5) + 1
	items := make([]domain.Item, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = generateRandomItem(trackNumber)
	}
	return domain.Order{
		OrderUID:          orderUID,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Delivery:          generateRandomDelivery(),
		Payment:           generateRandomPayment(orderUID, now),
		Items:             items,
		Locale:            generateLocale(),
		InternalSignature: "",
		CustomerID:        generateCustomerID(),
		DeliveryService:   generateDeliveryService(),
		Shardkey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              rand.Intn(100) + 1,
		DateCreated:       now,
		OofShard:          fmt.Sprintf("%d", rand.Intn(3)+1),
	}
}

func generateRandomDelivery() domain.Delivery {
	return domain.Delivery{
		Name:    gofakeit.Name(),
		Phone:   generatePhone(),
		Zip:     gofakeit.Zip(),
		City:    gofakeit.City(),
		Address: gofakeit.Street() + ", " + gofakeit.DigitN(2),
		Region:  gofakeit.State(),
		Email:   gofakeit.Email(),
	}
}

func generateRandomPayment(orderUID string, now time.Time) domain.Payment {
	amount := rand.Intn(20000) + 1000
	deliveryCost := rand.Intn(5000) + 500
	goodsTotal := amount - deliveryCost
	if goodsTotal < 100 {
		goodsTotal = amount - 500
	}
	return domain.Payment{
		Transaction:  orderUID,
		RequestID:    uuid.New().String(),
		Currency:     "RUB",
		Provider:     generatePaymentProvider(),
		Amount:       amount,
		PaymentDt:    now.Unix() - int64(rand.Intn(3600)),
		Bank:         generateBank(),
		DeliveryCost: deliveryCost,
		GoodsTotal:   goodsTotal,
		CustomFee:    rand.Intn(100),
	}
}

func generateRandomItem(trackNumber string) domain.Item {
	price := rand.Intn(5000) + 100
	quantity := rand.Intn(5) + 1
	sale := rand.Intn(50)
	return domain.Item{
		ChrtID:      rand.Intn(10000000),
		TrackNumber: trackNumber,
		Price:       price,
		Rid:         uuid.New().String(),
		Name:        generateProductName(),
		Sale:        sale,
		Size:        generateSize(),
		TotalPrice:  price * quantity * (100 - sale) / 100,
		NmID:        rand.Intn(3000000),
		Brand:       generateBrand(),
		Status:      generateItemStatus(),
	}
}

func generatePhone() string {
	format := "+7%s%s%s%s%s%s%s%s%s%s"
	return fmt.Sprintf(format,
		gofakeit.Digit(), gofakeit.Digit(), gofakeit.Digit(),
		gofakeit.Digit(), gofakeit.Digit(), gofakeit.Digit(),
		gofakeit.Digit(), gofakeit.Digit(), gofakeit.Digit(),
		gofakeit.Digit())
}

func generatePaymentProvider() string {
	providers := []string{"wbpay", "stripe", "paypal", "yoomoney", "tinkoff", "sberbank"}
	return providers[rand.Intn(len(providers))]
}

func generateBank() string {
	banks := []string{"alpha", "sber", "tinkoff", "vtb", "gazprom", "raiffeisen"}
	return banks[rand.Intn(len(banks))]
}

func generateProductName() string {
	products := []string{
		"Mascaras", "Lipstick", "Foundation", "Eyeshadow", "Blush",
		"Concealer", "Highlighter", "Bronzer", "Setting Spray", "Makeup Brushes",
		"Skincare Set", "Perfume", "Shampoo", "Conditioner", "Hair Mask",
		"Face Cream", "Serum", "Toner", "Cleanser", "Sunscreen",
	}
	return products[rand.Intn(len(products))]
}

func generateBrand() string {
	brands := []string{
		"Vivienne Sabo", "L'Oreal", "Maybelline", "MAC", "NYX",
		"Estee Lauder", "Clinique", "Chanel", "Dior", "YSL",
		"NARS", "Urban Decay", "Too Faced", "Fenty Beauty", "Huda Beauty",
		"Revlon", "Max Factor", "Bourjois", "Lancome", "Guerlain",
	}
	return brands[rand.Intn(len(brands))]
}

func generateSize() string {
	sizes := []string{"0", "XS", "S", "M", "L", "XL", "XXL", "36", "38", "40", "42", "44"}
	return sizes[rand.Intn(len(sizes))]
}

func generateItemStatus() int {
	statuses := []int{200, 201, 202, 300, 301, 400, 404, 500}
	return statuses[rand.Intn(len(statuses))]
}

func generateLocale() string {
	locales := []string{"en", "ru", "de", "fr", "es", "it", "zh", "ja", "ko"}
	return locales[rand.Intn(len(locales))]
}

func generateCustomerID() string {
	prefixes := []string{"user", "customer", "client", "buyer"}
	return fmt.Sprintf("%s_%s", prefixes[rand.Intn(len(prefixes))], gofakeit.DigitN(6))
}

func generateDeliveryService() string {
	services := []string{"meest", "russianpost", "dhl", "fedex", "ups", "cdek", "boxberry"}
	return services[rand.Intn(len(services))]
}
