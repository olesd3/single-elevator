package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func trackPosition(drv_floors2 chan int, d *elevio.MotorDirection) {
	for {
		a := <-drv_floors2
		// Even indices are floors, odd indices are in-between floors
		// Get the current floor
		mutex_d.Lock()
		currentFloor := 0
		for i := 0; i < 2*numFloors-1; i++ {
			if posArray[i] {
				currentFloor = i
			}
		}

		if a == -1 {
			if *d == elevio.MD_Up {
				mutex_posArray.Lock()
				posArray[currentFloor] = false
				posArray[currentFloor+1] = true
				mutex_posArray.Unlock()
			}
			if *d == elevio.MD_Down {
				mutex_posArray.Lock()
				posArray[currentFloor] = false
				posArray[currentFloor-1] = true
				mutex_posArray.Unlock()
			}
		} else {
			mutex_posArray.Lock()
			posArray[currentFloor] = false
			posArray[a*2] = true
			mutex_posArray.Unlock()
		}

		//fmt.Printf("Position array: %+v\n", posArray)
		mutex_d.Unlock()
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

func printElevatorOrders(elevatorOrders []Order) {
	printOrders(elevatorOrders)
}

func reverseDirection(d *elevio.MotorDirection) {
	switch {
	case *d == elevio.MD_Down:
		*d = elevio.MD_Up
	case *d == elevio.MD_Up:
		*d = elevio.MD_Down
	case *d == elevio.MD_Stop:
	}
}

// This function is only used internally in the sorting functions
func updatePosArray(dir elevio.MotorDirection, posArray *[2*numFloors - 1]bool) {
	// Reset all values in the array to false
	for i := range posArray {
		(posArray[i]) = false
	}

	switch {
	case dir == elevio.MD_Down:
		posArray[2*numFloors-2] = true
	case dir == elevio.MD_Up:
		posArray[0] = true
	default:
		panic("Error: MotorDirection MD_Stop passed into updatePosArray function")
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



// This function deletes relevant orders at the same floor as the current order, 
// It takes into account if there are multiple orders to the same floor
// Since elevatorOrders is sorted, we can just delete from left to right until there are no orders with the same floor left
func PopOrders() { 
	//fmt.Printf("Before deleting orders from elevatorOrders: %v\n", elevatorOrders)
	if len(elevatorOrders) != 0 {
		floor_to_pop := elevatorOrders[0].floor

		// Figure out how many elements to delete
		ndelete := 0
		for _, order := range elevatorOrders {
			if order.floor == floor_to_pop {
				ndelete += 1
			} else {
				break
			}
		}

		// Now that we've calculated the number of elements to delete, update elevatorOrders
		elevatorOrders = elevatorOrders[ndelete:]
	}
	//fmt.Printf("After deleting orders from elevatorOrders: %v\n", elevatorOrders)
}

func changeDirBasedOnCurrentOrder(d * elevio.MotorDirection, current_order Order, current_floor float32) {
	switch {
		case current_floor > float32(current_order.floor):
			*d = elevio.MD_Down
		case current_floor < float32(current_order.floor):
			*d = elevio.MD_Up
		case current_floor == float32(current_order.floor):
			*d = elevio.MD_Stop
	}
}

// This function will attend to the current order, it
func attendToSpecificOrder(d * elevio.MotorDirection, drv_floors chan int, drv_newOrder chan Order, posArray * [2*numFloors - 1]bool) {
	current_order := Order{0,-1,0}
	for {
		select {
			case a := <- drv_floors: 		                      // Triggers when we arrive at a new floor
				if a == current_order.floor {                     // Check if our new floor is equal to the floor of the order
                                        
					lockMutexes(&mutex_d, &mutex_elevatorOrders)  // Set direction to stop and delete relevant orders from elevatorOrders
					*d = elevio.MD_Stop
					elevio.SetMotorDirection(*d)
					PopOrders()
					unlockMutexes(&mutex_d, &mutex_elevatorOrders)
					
					
					time.Sleep(3000 * time.Millisecond)           // Wait for three seconds

					                                              // After deleting the relevant orders at our floor => find, if any, the next currentOrder
					lockMutexes(&mutex_d, &mutex_elevatorOrders)
					if len(elevatorOrders) != 0 { 	
						current_order = elevatorOrders[0]
						changeDirBasedOnCurrentOrder(d, current_order, float32(a))
						elevio.SetMotorDirection(*d)
					} 
					unlockMutexes(&mutex_d, &mutex_elevatorOrders)
				}
			case a := <- drv_newOrder:                 // If we get a new order => update current order and see if we need to redirect our elevator
				current_order = a
				lockMutexes(&mutex_d, &mutex_posArray)
				current_position := extractPos()
				changeDirBasedOnCurrentOrder(d, current_order, current_position)
				elevio.SetMotorDirection(*d)
				unlockMutexes(&mutex_d, &mutex_posArray)
		}
	}
}
