package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// ТЗ: Сервис для работы с файлами и базой данных.
// Требуется правильная обработка ошибок с сохранением контекста для debugging.

type FileService struct {
	db *sql.DB
}

func NewFileService(db *sql.DB) *FileService {
	return &FileService{db: db}
}

func (s *FileService) ProcessFile(filepath string) error {
	content, err := s.readFile(filepath)
	if err != nil {
		return err
	}

	metadata, err := s.extractMetadata(content)
	if err != nil {
		return err
	}

	err = s.saveToDatabase(filepath, metadata)
	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) readFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read file")
	}

	return content, nil
}

func (s *FileService) extractMetadata(content []byte) (map[string]interface{}, error) {
	var metadata map[string]interface{}
	err := json.Unmarshal(content, &metadata)
	if err != nil {
		return nil, errors.New("invalid JSON format")
	}

	if metadata["name"] == nil {
		return nil, errors.New("missing required field")
	}

	return metadata, nil
}

func (s *FileService) saveToDatabase(filepath string, metadata map[string]interface{}) error {
	name, ok := metadata["name"].(string)
	if !ok {
		return errors.New("invalid name type")
	}

	query := "INSERT INTO files (filepath, name) VALUES ($1, $2)"
	_, err := s.db.Exec(query, filepath, name)
	if err != nil {
		return errors.New("database error")
	}

	return nil
}

func (s *FileService) GetFileInfo(id int) (map[string]interface{}, error) {
	row := s.db.QueryRow("SELECT id, filepath, name FROM files WHERE id = $1", id)

	var fileID int
	var filepath, name string
	err := row.Scan(&fileID, &filepath, &name)
	if err != nil {
		return nil, errors.New("file not found")
	}

	return map[string]interface{}{
		"id":       fileID,
		"filepath": filepath,
		"name":     name,
	}, nil
}

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *APIClient) FetchData(endpoint string) ([]byte, error) {
	url := c.baseURL + endpoint

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, errors.New("request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("bad status code")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response")
	}

	return body, nil
}

func (c *APIClient) ProcessAndStore(endpoint string, service *FileService) error {
	data, err := c.FetchData(endpoint)
	if err != nil {
		return err
	}

	var items []map[string]interface{}
	err = json.Unmarshal(data, &items)
	if err != nil {
		return errors.New("invalid response format")
	}

	for _, item := range items {
		filepath := fmt.Sprintf("/tmp/%v.json", item["id"])

		file, err := os.Create(filepath)
		if err != nil {
			return errors.New("failed to create file")
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			file.Close()
			return errors.New("failed to marshal item")
		}

		_, err = file.Write(jsonData)
		file.Close()
		if err != nil {
			return errors.New("failed to write file")
		}

		err = service.ProcessFile(filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	db, err := sql.Open("postgres", "postgresql://user:pass@localhost/testdb")
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}
	defer db.Close()

	service := NewFileService(db)

	err = service.ProcessFile("/path/to/file.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	client := NewAPIClient("https://api.example.com")
	err = client.ProcessAndStore("/data", service)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Success!")
}
