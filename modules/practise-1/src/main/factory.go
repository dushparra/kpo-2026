package main

type HseCarFactory struct {
	nextVin   int
	cars      []*Car
	customers []*Customer
}

func NewHseCarFactory() *HseCarFactory {
	return &HseCarFactory{
		nextVin: 1,
	}
}

func (f *HseCarFactory) AddCar(pedalSize int) {
	car := &Car{
		VIN: f.nextVin,
		Engine: Engine{
			PedalSize: pedalSize,
		},
	}

	f.cars = append(f.cars, car)
	f.nextVin++
}

func (f *HseCarFactory) AddCustomer(customer *Customer) {
	f.customers = append(f.customers, customer)
}

func (f *HseCarFactory) SellCars() {
	freeCarInd := 0

	for _, customer := range f.customers {
		if customer.Car != nil {
			continue
		}

		if freeCarInd >= len(f.cars) {
			break
		}

		customer.Car = f.cars[freeCarInd]
		freeCarInd++
	}

	f.cars = nil
}
