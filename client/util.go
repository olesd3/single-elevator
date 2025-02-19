package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func trackPosition(drv_floors2 chan int, d *elevio.MotorDirection) {
	for {
		select {
		case a := <-drv_floors2:
			// Even indices are floors, odd indices are in-between floors
			// Get the current floor
			currentFloor := 0
			for i := 0; i < 2*numFloors-1; i++ {
				if posArray[i] {
					currentFloor = i
				}
			}

			if a == -1 {
				if *d == elevio.MD_Up {
					mu2.Lock()
					posArray[currentFloor] = false
					posArray[currentFloor+1] = true
					mu2.Unlock()
				}
				if *d == elevio.MD_Down {
					mu2.Lock()
					posArray[currentFloor] = false
					posArray[currentFloor-1] = true
					mu2.Unlock()
				}
			} else {
				mu2.Lock()
				posArray[currentFloor] = false
				posArray[a*2] = true
				mu2.Unlock()
			}

			fmt.Printf("Position array: %+v\n", posArray)
		}
	}
}

func addOrder(floor int, direction OrderDirection, typeOrder OrderType) {
	exists := false

	if typeOrder == cab {
		for _, order := range elevatorOrders {
			if order.floor == floor && order.orderType == cab {
				exists = true
			}
		}
	} else if typeOrder == hall {
		for _, order := range elevatorOrders {
			if order.floor == floor && order.direction == direction && order.orderType == hall {
				exists = true
			}
		}
	}

	if !exists {
		elevatorOrders = append(elevatorOrders, Order{floor: floor, direction: direction, orderType: typeOrder})
	}
}

// Helper function to convert direction to string
func directionToString(direction OrderDirection) string {
	if direction == up {
		return "Up"
	}
	return "Down"
}

// Helper function to convert orderType to string
func orderTypeToString(orderType OrderType) string {
	if orderType == hall {
		return "Hall"
	}
	return "Cab"
}

func printOrders(orders []Order) {
	// Iterate through each order and print its details
	for i, order := range orders {
		fmt.Printf("Order #%d: Floor %d, Direction %s, OrderType %s\n",
			i+1, order.floor, directionToString(order.direction), orderTypeToString(order.orderType))
	}
}

func reverseDirection(d *MotorDirection) {
	switch {
	case *d == MD_Down:
		*d = MD_Up
	case *d == MD_Up:
		*d = MD_Down
	case *d == MD_Stop:
	}
}

func updatePosArray(dir MotorDirection, posArray *[2*numFloors - 1]bool) {
	// Reset all values in the array to false
	for i := range posArray {
		(posArray[i]) = false
	}

	switch {
	case dir == MD_Down:
		posArray[2*numFloors-2] = true
	case dir == MD_Up:
		posArray[0] = true
	default:
		panic("Error: MotorDirection MD_Stop passed into updatePosArray function")
	}
}

func EventDirStop(receiver chan<- MotorDirection) {
	prev := lastDir

	direction := newDir

	if direction != prev {
		receiver <- direction
	}
}

func handleNewOrder(newOrderTrigger chan bool) {
	for {
		//Blocks until we receive a new order

		received_order := <-newOrderTrigger
		// If we get a new order sort elevatorOrders
		if received_order {
			mu1.Lock()
			sortAllOrders(&elevatorOrders, direction, posArray)
			mu1.Unlock()
		}
		currentOrder = elevatorOrders[0]
	}
}

func extractPos() float32 {
	currentFloor := float32(0)
	for i := 0; i < 2*numFloors-1; i++ {
		if posArray[i] {
			currentFloor = float32(i) / 2
		}
	}
	return currentFloor
}

// Global: currentOrder,   Local: local_currentOrder
func AttendToSpecificOrder(d *elevio.MotorDirection) {
	// Defining a boolean value that dictates whether or not we have a current order
	haveCurrentOrder := false
	local_currentOrder := Order{floor: 0, direction: down, orderType: hall}

	for {
		switch {
		case !haveCurrentOrder:
			// Case 1: No current order exists:
			blocker := <-receiver
			if blocker {
				haveCurrentOrder = true
			}

		case haveCurrentOrder:
			haveCompletedOrder := false
			for !haveCompletedOrder {
				switch {
				case float32(currentOrder.floor) > extractPos():
					elevio.SetMotorDirection(elevio.MD_Up)
				case float32(currentOrder.floor) < extractPos():
					elevio.SetMotorDirection(elevio.MD_Down)
				case float32(currentOrder.floor) == extractPos():
					elevio.SetMotorDirection(elevio.MD_Stop)

					// Popping the order from elevatorOrders
					// Mutex lock
					time.Sleep(3000 * time.Millisecond)
					// Mutex unlock

					// Sort the elevatorOrders for safety sake

				}
			}
			// Case 2: Current order exists
			/*
				local_currentOrder = currentOrder
				for local_currentOrder == currentOrder {

					current_elevator_position := extractPos()
					order_floor := float32(currentOrder.floor)

					switch {
						case order_floor > current_elevator_position:
							elevio.SetMotorDirection(elevio.MD_Up)
						case order_floor < current_elevator_position:
							elevio.SetMotorDirection(elevio.MD_Down)
						case order_floor == current_elevator_position:
							elevio.SetMotorDirection(elevio.MD_Stop)

							// If it hasnt change for a while, we're idle
							time.Sleep(3000 * time.Millisecond)
							// TODO: opens doors, door lights, close doors

							// Find local_currentOrder in elevatorOrders and remove it

							// Consider the rest of elevatorOrders
							// If there are more orders:
							//     set local to currentOrder
							// Else if no more orders:
							//     set haveCurrentOrder to false
					}

					}

				}
			*/

			// While (Current order hasn't changed) => GoTo that floor

			// When done with the order => remove it from elevatorOrders

			// Now that we've popped the order, set, if there still are elements in elevatorOrders, set the new current order to the first element in elevatorOrders

		}
	}

}
