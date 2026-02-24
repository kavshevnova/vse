package main

// Задача: Validator — композитный валидатор с несколькими стратегиями.

import (
	"fmt"
	"regexp"
	"sync"
)

type ValidationError struct {
	Field   string
	Message string
}

type Validator interface {
	Validate(data interface{}) []ValidationError
}

type CompositeValidator interface {
	Validator
	Add(validator Validator)
	ValidateField(field string, value interface{}) []ValidationError
}

// --- RequiredValidator ---

type RequiredValidator struct{ Field string }

func (v *RequiredValidator) Validate(data interface{}) []ValidationError {
	if data == nil || data == "" {
		return []ValidationError{{Field: v.Field, Message: "required"}}
	}
	return nil
}

// --- LengthValidator ---

type LengthValidator struct {
	Field string
	Min   int
	Max   int
}

func (v *LengthValidator) Validate(data interface{}) []ValidationError {
	s, ok := data.(string)
	if !ok {
		return nil
	}
	if len(s) < v.Min || (v.Max > 0 && len(s) > v.Max) {
		return []ValidationError{{Field: v.Field, Message: fmt.Sprintf("length must be between %d and %d", v.Min, v.Max)}}
	}
	return nil
}

// --- EmailValidator ---

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type EmailValidator struct{ Field string }

func (v *EmailValidator) Validate(data interface{}) []ValidationError {
	s, ok := data.(string)
	if !ok || !emailRe.MatchString(s) {
		return []ValidationError{{Field: v.Field, Message: "invalid email format"}}
	}
	return nil
}

// --- ChainValidator (последовательно) ---

type ChainValidator struct {
	validators []Validator
}

func (c *ChainValidator) Add(v Validator)                               { c.validators = append(c.validators, v) }
func (c *ChainValidator) ValidateField(f string, val interface{}) []ValidationError { return c.Validate(val) }

func (c *ChainValidator) Validate(data interface{}) []ValidationError {
	for _, v := range c.validators {
		if errs := v.Validate(data); len(errs) > 0 {
			return errs // останавливаемся на первой ошибке
		}
	}
	return nil
}

// --- ParallelValidator (параллельно) ---

type ParallelValidator struct {
	validators []Validator
}

func (p *ParallelValidator) Add(v Validator)                               { p.validators = append(p.validators, v) }
func (p *ParallelValidator) ValidateField(f string, val interface{}) []ValidationError { return p.Validate(val) }

func (p *ParallelValidator) Validate(data interface{}) []ValidationError {
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []ValidationError
	)
	for _, v := range p.validators {
		wg.Add(1)
		go func(v Validator) {
			defer wg.Done()
			if e := v.Validate(data); len(e) > 0 {
				mu.Lock()
				errs = append(errs, e...)
				mu.Unlock()
			}
		}(v)
	}
	wg.Wait()
	return errs
}

func main() {
	chain := &ChainValidator{}
	chain.Add(&RequiredValidator{Field: "email"})
	chain.Add(&LengthValidator{Field: "email", Min: 5, Max: 100})
	chain.Add(&EmailValidator{Field: "email"})

	fmt.Println(chain.Validate(""))            // required
	fmt.Println(chain.Validate("ab"))          // too short
	fmt.Println(chain.Validate("notanemail"))  // invalid email
	fmt.Println(chain.Validate("a@b.com"))     // nil

	parallel := &ParallelValidator{}
	parallel.Add(&RequiredValidator{Field: "name"})
	parallel.Add(&LengthValidator{Field: "name", Min: 2, Max: 50})
	fmt.Println(parallel.Validate("")) // оба валидатора сработают
}
