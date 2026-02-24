package main

import (
	"errors"
	"math"
)

type parrotType int

const (
	TypeEuropean      parrotType = 1 // iota instead of nums
	TypeAfrican       parrotType = 2
	TypeNorwegianBlue parrotType = 3
)

type Parrot interface {
	Speed() (float64, error)
}

type mixedParrot struct {
	_type            parrotType // type instead of _type
	numberOfCoconuts int
	voltage          float64
	nailed           bool
}

func CreateParrot(t parrotType, numberOfCoconuts int, voltage float64, nailed bool) Parrot {
	return &mixedParrot{t, numberOfCoconuts, voltage, nailed}
}

func (parrot mixedParrot) Speed() (float64, error) { // open closed principle
	switch parrot._type {
	case TypeEuropean:
		return parrot.baseSpeed(), nil

	case TypeAfrican:
		return math.Max(0, parrot.baseSpeed()-parrot.loadFactor()*float64(parrot.numberOfCoconuts)), nil

	case TypeNorwegianBlue:
		if parrot.nailed {
			return 0, nil
		}
		return parrot.computeBaseSpeedForVoltage(parrot.voltage), nil

	default:
		return 0, errors.New("should be unreachable")
	}
}

func (parrot mixedParrot) computeBaseSpeedForVoltage(voltage float64) float64 {
	return math.Min(24.0, voltage*parrot.baseSpeed())
}

func (parrot mixedParrot) loadFactor() float64 {
	return 9.0 // moved to const
}

func (parrot mixedParrot) baseSpeed() float64 {
	return 12.0
}
