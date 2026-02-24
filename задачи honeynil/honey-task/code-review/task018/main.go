package main

import (
	"errors"
	"fmt"
	"strings"
)

// ТЗ: Модуль для валидации и обработки пользовательских данных.
// Код должен иметь хорошее покрытие тестами и быть легко тестируемым.
// Найти проблемы, которые затрудняют тестирование кода.

type UserValidator struct {
	minPasswordLength int
	requiredDomains   []string
}

func NewUserValidator() *UserValidator {
	return &UserValidator{
		minPasswordLength: 8,
		requiredDomains:   []string{"example.com", "company.com"},
	}
}

func (v *UserValidator) ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username too short")
	}

	if len(username) > 20 {
		return errors.New("username too long")
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}

func (v *UserValidator) ValidateEmail(email string) error {
	if !strings.Contains(email, "@") {
		return errors.New("invalid email format")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return errors.New("invalid email format")
	}

	validDomain := false
	for _, domain := range v.requiredDomains {
		if parts[1] == domain {
			validDomain = true
			break
		}
	}

	if !validDomain {
		return errors.New("email domain not allowed")
	}

	return nil
}

func (v *UserValidator) ValidatePassword(password string) error {
	if len(password) < v.minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", v.minPasswordLength)
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= 'a' && char <= 'z' {
			hasLower = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain digit")
	}

	return nil
}

type UserService struct {
	validator *UserValidator
	users     map[string]User
}

type User struct {
	Username string
	Email    string
	Password string
}

func NewUserService() *UserService {
	return &UserService{
		validator: NewUserValidator(),
		users:     make(map[string]User),
	}
}

func (s *UserService) RegisterUser(username, email, password string) error {
	if err := s.validator.ValidateUsername(username); err != nil {
		return err
	}

	if err := s.validator.ValidateEmail(email); err != nil {
		return err
	}

	if err := s.validator.ValidatePassword(password); err != nil {
		return err
	}

	if _, exists := s.users[username]; exists {
		return errors.New("username already exists")
	}

	s.users[username] = User{
		Username: username,
		Email:    email,
		Password: password,
	}

	fmt.Printf("User %s registered successfully\n", username)
	return nil
}

func (s *UserService) GetUser(username string) (User, error) {
	user, exists := s.users[username]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateEmail(username, newEmail string) error {
	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}

	if err := s.validator.ValidateEmail(newEmail); err != nil {
		return err
	}

	user.Email = newEmail
	s.users[username] = user

	fmt.Printf("Email updated for user %s\n", username)
	return nil
}

func (s *UserService) ChangePassword(username, oldPassword, newPassword string) error {
	user, exists := s.users[username]
	if !exists {
		return errors.New("user not found")
	}

	if user.Password != oldPassword {
		return errors.New("incorrect old password")
	}

	if err := s.validator.ValidatePassword(newPassword); err != nil {
		return err
	}

	user.Password = newPassword
	s.users[username] = user

	fmt.Printf("Password changed for user %s\n", username)
	return nil
}

func (s *UserService) DeleteUser(username string) error {
	if _, exists := s.users[username]; !exists {
		return errors.New("user not found")
	}

	delete(s.users, username)
	fmt.Printf("User %s deleted\n", username)
	return nil
}

func main() {
	service := NewUserService()

	err := service.RegisterUser("john_doe", "john@example.com", "Password123")
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	}

	err = service.RegisterUser("invalid user", "invalid@wrong.com", "weak")
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	}

	user, err := service.GetUser("john_doe")
	if err == nil {
		fmt.Printf("Found user: %+v\n", user)
	}

	err = service.UpdateEmail("john_doe", "john.doe@company.com")
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
	}

	err = service.ChangePassword("john_doe", "Password123", "NewPassword456")
	if err != nil {
		fmt.Printf("Password change failed: %v\n", err)
	}
}
