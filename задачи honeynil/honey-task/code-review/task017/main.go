package main

import (
	"fmt"
	"strings"
)

// ТЗ: Генератор HTML-страниц и отчетов на основе данных.
// Необходимо оптимизировать работу со строками для больших объемов данных.

type User struct {
	ID       int
	Username string
	Email    string
	Role     string
}

type Report struct {
	Title   string
	Date    string
	Content string
}

func GenerateHTMLTable(users []User) string {
	html := "<table>"
	html += "<thead>"
	html += "<tr>"
	html += "<th>ID</th>"
	html += "<th>Username</th>"
	html += "<th>Email</th>"
	html += "<th>Role</th>"
	html += "</tr>"
	html += "</thead>"
	html += "<tbody>"

	for _, user := range users {
		html += "<tr>"
		html += "<td>" + fmt.Sprintf("%d", user.ID) + "</td>"
		html += "<td>" + user.Username + "</td>"
		html += "<td>" + user.Email + "</td>"
		html += "<td>" + user.Role + "</td>"
		html += "</tr>"
	}

	html += "</tbody>"
	html += "</table>"

	return html
}

func GenerateCSV(users []User) string {
	csv := "ID,Username,Email,Role\n"

	for _, user := range users {
		csv += fmt.Sprintf("%d", user.ID) + ","
		csv += user.Username + ","
		csv += user.Email + ","
		csv += user.Role + "\n"
	}

	return csv
}

func BuildSQLInserts(users []User) string {
	sql := ""

	for _, user := range users {
		sql += "INSERT INTO users (id, username, email, role) VALUES ("
		sql += fmt.Sprintf("%d", user.ID) + ", "
		sql += "'" + user.Username + "', "
		sql += "'" + user.Email + "', "
		sql += "'" + user.Role + "');\n"
	}

	return sql
}

func GenerateReport(reports []Report) string {
	output := "=" + strings.Repeat("=", 50) + "\n"
	output += "MONTHLY REPORTS\n"
	output += "=" + strings.Repeat("=", 50) + "\n\n"

	for i, report := range reports {
		output += fmt.Sprintf("Report #%d\n", i+1)
		output += "-" + strings.Repeat("-", 50) + "\n"
		output += "Title: " + report.Title + "\n"
		output += "Date: " + report.Date + "\n"
		output += "-" + strings.Repeat("-", 50) + "\n"
		output += report.Content + "\n\n"
	}

	output += "=" + strings.Repeat("=", 50) + "\n"
	output += fmt.Sprintf("Total Reports: %d\n", len(reports))
	output += "=" + strings.Repeat("=", 50) + "\n"

	return output
}

func BuildJSONArray(users []User) string {
	json := "[\n"

	for i, user := range users {
		json += "  {\n"
		json += fmt.Sprintf("    \"id\": %d,\n", user.ID)
		json += "    \"username\": \"" + user.Username + "\",\n"
		json += "    \"email\": \"" + user.Email + "\",\n"
		json += "    \"role\": \"" + user.Role + "\"\n"
		json += "  }"

		if i < len(users)-1 {
			json += ","
		}
		json += "\n"
	}

	json += "]"
	return json
}

func GenerateEmailBody(user User, subject string, message string) string {
	email := "To: " + user.Email + "\n"
	email += "From: noreply@example.com\n"
	email += "Subject: " + subject + "\n\n"
	email += "Dear " + user.Username + ",\n\n"
	email += message + "\n\n"
	email += "Best regards,\n"
	email += "The Team\n"
	email += "---\n"
	email += "This is an automated message. Please do not reply.\n"

	return email
}

func CreateLogMessage(level string, service string, message string, details map[string]string) string {
	log := "[" + level + "] "
	log += service + ": "
	log += message

	if len(details) > 0 {
		log += " | "
		for key, value := range details {
			log += key + "=" + value + " "
		}
	}

	return log
}

func main() {
	// Генерируем тестовые данные
	users := make([]User, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = User{
			ID:       i + 1,
			Username: fmt.Sprintf("user%d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Role:     "member",
		}
	}

	fmt.Println("Generating HTML table...")
	html := GenerateHTMLTable(users)
	fmt.Printf("HTML length: %d characters\n", len(html))

	fmt.Println("Generating CSV...")
	csv := GenerateCSV(users)
	fmt.Printf("CSV length: %d characters\n", len(csv))

	fmt.Println("Generating SQL inserts...")
	sql := BuildSQLInserts(users)
	fmt.Printf("SQL length: %d characters\n", len(sql))

	reports := make([]Report, 100)
	for i := 0; i < 100; i++ {
		reports[i] = Report{
			Title:   fmt.Sprintf("Report %d", i+1),
			Date:    "2024-01-15",
			Content: "This is the content of report number " + fmt.Sprintf("%d", i+1),
		}
	}

	fmt.Println("Generating report...")
	report := GenerateReport(reports)
	fmt.Printf("Report length: %d characters\n", len(report))

	fmt.Println("Generating JSON...")
	json := BuildJSONArray(users[:100])
	fmt.Printf("JSON length: %d characters\n", len(json))

	fmt.Println("Generating email...")
	email := GenerateEmailBody(users[0], "Welcome!", "Thank you for joining our platform.")
	fmt.Printf("Email length: %d characters\n", len(email))

	details := map[string]string{
		"user_id":   "123",
		"action":    "login",
		"ip":        "192.168.1.1",
		"timestamp": "2024-01-15T10:30:00Z",
	}
	log := CreateLogMessage("INFO", "AuthService", "User logged in successfully", details)
	fmt.Println(log)
}
