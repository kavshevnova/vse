package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ТЗ: API для работы с различными типами данных и динамической обработкой JSON.
// Необходимо безопасно работать с интерфейсами и type assertions.

type Config struct {
	Settings map[string]interface{}
}

var globalConfig *Config

func InitConfig(data map[string]interface{}) {
	globalConfig = &Config{
		Settings: data,
	}
}

func GetStringSetting(key string) string {
	value := globalConfig.Settings[key]
	return value.(string)
}

func GetIntSetting(key string) int {
	value := globalConfig.Settings[key]
	return value.(int)
}

func GetBoolSetting(key string) bool {
	value := globalConfig.Settings[key]
	return value.(bool)
}

type DataProcessor struct {
	transformer func(interface{}) interface{}
}

func NewDataProcessor(transformer func(interface{}) interface{}) *DataProcessor {
	return &DataProcessor{
		transformer: transformer,
	}
}

func (p *DataProcessor) Process(data interface{}) interface{} {
	return p.transformer(data)
}

type Response struct {
	Data     interface{}
	Error    string
	Metadata map[string]interface{}
}

func handleDynamicData(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	json.NewDecoder(r.Body).Decode(&input)

	dataType := input["type"].(string)
	payload := input["payload"]

	var result interface{}

	switch dataType {
	case "user":
		user := payload.(map[string]interface{})
		result = map[string]interface{}{
			"id":   user["id"].(int),
			"name": user["name"].(string),
			"age":  user["age"].(float64),
		}

	case "product":
		product := payload.(map[string]interface{})
		result = map[string]interface{}{
			"id":    product["id"].(int),
			"title": product["title"].(string),
			"price": product["price"].(float64),
		}

	case "order":
		order := payload.(map[string]interface{})
		items := order["items"].([]interface{})

		total := 0.0
		for _, item := range items {
			itemMap := item.(map[string]interface{})
			total += itemMap["price"].(float64) * itemMap["quantity"].(float64)
		}

		result = map[string]interface{}{
			"order_id": order["id"].(int),
			"total":    total,
			"items":    len(items),
		}
	}

	response := Response{
		Data: result,
	}

	json.NewEncoder(w).Encode(response)
}

func calculateTotal(items []interface{}) float64 {
	total := 0.0

	for _, item := range items {
		itemData := item.(map[string]interface{})
		price := itemData["price"].(float64)
		quantity := itemData["quantity"].(int)
		total += price * float64(quantity)
	}

	return total
}

type UserProfile struct {
	ID       int
	Username string
	Email    *string
	Phone    *string
	Settings *Config
}

func (u *UserProfile) GetEmail() string {
	return *u.Email
}

func (u *UserProfile) GetPhone() string {
	return *u.Phone
}

func (u *UserProfile) GetSetting(key string) interface{} {
	return u.Settings.Settings[key]
}

func (u *UserProfile) UpdateSettings(updates map[string]interface{}) {
	for key, value := range updates {
		u.Settings.Settings[key] = value
	}
}

func processUserProfile(profile *UserProfile) string {
	info := fmt.Sprintf("User: %s\n", profile.Username)
	info += fmt.Sprintf("Email: %s\n", profile.GetEmail())
	info += fmt.Sprintf("Phone: %s\n", profile.GetPhone())

	theme := profile.GetSetting("theme").(string)
	notifications := profile.GetSetting("notifications").(bool)

	info += fmt.Sprintf("Theme: %s\n", theme)
	info += fmt.Sprintf("Notifications: %v\n", notifications)

	return info
}

func handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	var update map[string]interface{}
	json.NewDecoder(r.Body).Decode(&update)

	userId := update["user_id"].(int)

	profile := getUserProfile(userId)

	if email, ok := update["email"]; ok {
		emailStr := email.(string)
		profile.Email = &emailStr
	}

	if settings, ok := update["settings"]; ok {
		settingsMap := settings.(map[string]interface{})
		profile.UpdateSettings(settingsMap)
	}

	result := processUserProfile(profile)

	w.Write([]byte(result))
}

var userProfiles map[int]*UserProfile

func getUserProfile(id int) *UserProfile {
	return userProfiles[id]
}

func parseNestedData(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	dataMap := data.(map[string]interface{})

	for key, value := range dataMap {
		switch v := value.(type) {
		case map[string]interface{}:
			result[key] = parseNestedData(v)
		case []interface{}:
			var items []interface{}
			for _, item := range v {
				items = append(items, parseNestedData(item))
			}
			result[key] = items
		default:
			result[key] = v
		}
	}

	return result
}

type QueryBuilder struct {
	table   string
	fields  []string
	filters map[string]interface{}
}

func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.fields = fields
	return qb
}

func (qb *QueryBuilder) Where(key string, value interface{}) *QueryBuilder {
	qb.filters[key] = value
	return qb
}

func (qb *QueryBuilder) Build() string {
	query := "SELECT "

	for i, field := range qb.fields {
		query += field
		if i < len(qb.fields)-1 {
			query += ", "
		}
	}

	query += " FROM " + qb.table

	if len(qb.filters) > 0 {
		query += " WHERE "
		i := 0
		for key, value := range qb.filters {
			query += fmt.Sprintf("%s = %v", key, value)
			if i < len(qb.filters)-1 {
				query += " AND "
			}
			i++
		}
	}

	return query
}

func main() {
	configData := map[string]interface{}{
		"app_name": "MyApp",
		"port":     8080,
		"debug":    true,
	}
	InitConfig(configData)

	appName := GetStringSetting("app_name")
	port := GetIntSetting("port")
	debug := GetBoolSetting("debug")

	fmt.Printf("Starting %s on port %d (debug: %v)\n", appName, port, debug)

	userProfiles = make(map[int]*UserProfile)

	email := "user1@example.com"
	userProfiles[1] = &UserProfile{
		ID:       1,
		Username: "john_doe",
		Email:    &email,
		Settings: globalConfig,
	}

	http.HandleFunc("/api/dynamic", handleDynamicData)
	http.HandleFunc("/api/user/update", handleUserUpdate)

	fmt.Println("Server ready to handle requests")

	qb := &QueryBuilder{table: "users"}
	qb.Select("id", "name", "email").Where("status", "active")
	query := qb.Build()
	fmt.Println("Generated query:", query)

	timeout := GetIntSetting("timeout")
	fmt.Printf("Timeout: %d\n", timeout)
}
