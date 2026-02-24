package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ТЗ: REST API для управления товарами в интернет-магазине.
// Код должен корректно обрабатывать панику и возвращать понятные ошибки клиенту.
// Найти проблемы.
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var products = map[int]Product{
	1: {ID: 1, Name: "Laptop", Price: 999.99},
	2: {ID: 2, Name: "Mouse", Price: 29.99},
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	product := products[id]

	if product.ID == 0 {
		panic("Product not found!")
	}

	json.NewEncoder(w).Encode(product)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	json.NewDecoder(r.Body).Decode(&product)

	if product.Name == "" {
		panic("Product name is required")
	}

	if product.Price <= 0 {
		panic("Price must be positive")
	}

	product.ID = len(products) + 1
	products[product.ID] = product

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func calculateDiscount(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	discountStr := r.URL.Query().Get("discount")

	id, _ := strconv.Atoi(idStr)
	discount, _ := strconv.ParseFloat(discountStr, 64)

	product := products[id]
	finalPrice := product.Price * (1 - discount/100)

	response := map[string]interface{}{
		"original_price": product.Price,
		"discount":       discount,
		"final_price":    finalPrice,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/product", getProduct)
	http.HandleFunc("/product/create", createProduct)
	http.HandleFunc("/product/discount", calculateDiscount)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
