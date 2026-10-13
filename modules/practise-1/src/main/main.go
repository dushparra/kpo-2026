package main

import "fmt"

func main() {
	factory := NewHseCarFactory()

	factory.AddCar(1)
	factory.AddCar(2)
	factory.AddCar(3)
	factory.AddCar(4)

	factory.AddCustomer(&Customer{FIO: "Вася"})
	factory.AddCustomer(&Customer{FIO: "Вова"})
	factory.AddCustomer(&Customer{FIO: "Света"})

	fmt.Println("Автомобили до продажи")
	printCars(factory.cars)

	fmt.Println("Покупатели до продажи")
	printCustomers(factory.customers)

	factory.SellCars()

	fmt.Println("Автомобили после продажи")
	printCars(factory.cars)

	fmt.Println("Покупатели после продажи")
	printCustomers(factory.customers)
}

func printCars(cars []*Car) {
	if len(cars) == 0 {
		fmt.Println("На складе нет автомобилей")
		return
	}

	for _, car := range cars {
		fmt.Printf("VIN: %d, размер педалей: %d\n", car.VIN, car.Engine.PedalSize)
	}
}

func printCustomers(customers []*Customer) {
	for _, customer := range customers {
		if customer.Car == nil {
			fmt.Printf("Покупатель: %s, автомобиль: нет\n", customer.FIO)
			continue
		}

		fmt.Printf(
			"Покупатель: %s, VIN: %d, размер педалей: %d\n",
			customer.FIO,
			customer.Car.VIN,
			customer.Car.Engine.PedalSize,
		)
	}
}
