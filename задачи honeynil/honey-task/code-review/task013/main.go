package main

import (
	"fmt"
	"io"
)

// ТЗ: Разработать систему для работы с различными типами хранилищ данных.
// Код должен быть гибким и расширяемым.
// Найти проблемы.

// UserStorage определяет методы для работы с пользователями
type UserStorage interface {
	GetUserByID(id int) (User, error)
	SaveUser(user User) error
	DeleteUser(id int) error
	UpdateUser(user User) error
	ListUsers() ([]User, error)
	CountUsers() (int, error)
	UserExists(id int) (bool, error)
}

// ProductStorage определяет методы для работы с товарами
type ProductStorage interface {
	GetProductByID(id int) (Product, error)
	SaveProduct(product Product) error
	DeleteProduct(id int) error
	UpdateProduct(product Product) error
	ListProducts() ([]Product, error)
	CountProducts() (int, error)
	ProductExists(id int) (bool, error)
}

// OrderStorage определяет методы для работы с заказами
type OrderStorage interface {
	GetOrderByID(id int) (Order, error)
	SaveOrder(order Order) error
	DeleteOrder(id int) error
	UpdateOrder(order Order) error
	ListOrders() ([]Order, error)
	CountOrders() (int, error)
	OrderExists(id int) (bool, error)
}

// DataReader для чтения данных
type DataReader interface {
	Read() ([]byte, error)
	ReadAll() ([]byte, error)
	ReadLine() (string, error)
}

// DataWriter для записи данных
type DataWriter interface {
	Write(data []byte) error
	WriteLine(line string) error
	WriteAll(data []byte) error
}

// DataProcessor для обработки данных
type DataProcessor interface {
	Process(data []byte) ([]byte, error)
}

// Logger для логирования
type Logger interface {
	Log(message string)
	LogError(err error)
	LogInfo(message string)
	LogWarning(message string)
	LogDebug(message string)
}

type User struct {
	ID   int
	Name string
}

type Product struct {
	ID    int
	Title string
	Price float64
}

type Order struct {
	ID     int
	UserID int
	Total  float64
}

// MemoryUserStorage - реализация в памяти
type MemoryUserStorage struct {
	users map[int]User
}

func NewMemoryUserStorage() UserStorage {
	return &MemoryUserStorage{
		users: make(map[int]User),
	}
}

func (s *MemoryUserStorage) GetUserByID(id int) (User, error) {
	user, ok := s.users[id]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *MemoryUserStorage) SaveUser(user User) error {
	s.users[user.ID] = user
	return nil
}

func (s *MemoryUserStorage) DeleteUser(id int) error {
	delete(s.users, id)
	return nil
}

func (s *MemoryUserStorage) UpdateUser(user User) error {
	s.users[user.ID] = user
	return nil
}

func (s *MemoryUserStorage) ListUsers() ([]User, error) {
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

func (s *MemoryUserStorage) CountUsers() (int, error) {
	return len(s.users), nil
}

func (s *MemoryUserStorage) UserExists(id int) (bool, error) {
	_, ok := s.users[id]
	return ok, nil
}

// FileDataReader реализует чтение из файла
type FileDataReader struct {
	reader io.Reader
}

func NewFileDataReader(r io.Reader) DataReader {
	return &FileDataReader{reader: r}
}

func (f *FileDataReader) Read() ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := f.reader.Read(buf)
	return buf[:n], err
}

func (f *FileDataReader) ReadAll() ([]byte, error) {
	return io.ReadAll(f.reader)
}

func (f *FileDataReader) ReadLine() (string, error) {
	data, err := f.ReadAll()
	return string(data), err
}

func main() {
	storage := NewMemoryUserStorage()

	user := User{ID: 1, Name: "John"}
	storage.SaveUser(user)

	retrieved, _ := storage.GetUserByID(1)
	fmt.Printf("Retrieved user: %+v\n", retrieved)
}
