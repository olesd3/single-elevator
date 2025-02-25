package main

import (
	"Driver-go/elevio"
	"fmt"
	"sync"
)

const numFloors = 4

var (
	elevatorOrders       []Order
	mutex_elevatorOrders sync.Mutex
)

var (
	posArray       [2*numFloors - 1]bool
	mutex_posArray sync.Mutex
)

var mutex_d sync.Mutex

func lockMutexes(mutexes ...*sync.Mutex) {
	for _, m := range mutexes {
		m.Lock()
	}
}

func unlockMutexes(mutexes ...*sync.Mutex) {
	for _, m := range mutexes {
		m.Unlock()
	}
}

func main() {

	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_floors2 := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_newOrder := make(chan Order)
	drv_finishedInitialization := make(chan bool)

	go elevio.PollButtons(drv_buttons)         // Starts checking for button updates
	go elevio.PollFloorSensor(drv_floors)      // Starts checking for floors updates
	go elevio.PollFloorSensor2(drv_floors2)    // Starts checking for floors updates (for tracking position)
	go elevio.PollObstructionSwitch(drv_obstr) // Starts checking for obstruction updates
	go elevio.PollStopButton(drv_stop)         // Starts checking for stop button presses

	var d elevio.MotorDirection = elevio.MD_Down

	// Section_START ---- Initialization

	go func() {
		elevio.SetMotorDirection(d)
		for {
			a := <-drv_floors
			if a == 0 {
				d = elevio.MD_Stop
				elevio.SetMotorDirection(d)
				break
			}
		}
		drv_finishedInitialization <- true
	}()

	<-drv_finishedInitialization

	fmt.Printf("Initialization finished\n")

	// Section_END ---- Initialization

	go trackPosition(drv_floors2, &d) // Starts tracking the position of the elevator
	go attendToSpecificOrder(&d, drv_floors, drv_newOrder)

	for {
		select {
		case a := <-drv_buttons:
			// Gets a new order
			// Adds it to elevatorOrders and sorts
			lockMutexes(&mutex_elevatorOrders, &mutex_d, &mutex_posArray)

			switch {
			case a.Button == elevio.BT_HallUp:
				addOrder(a.Floor, up, hall)
			case a.Button == elevio.BT_HallDown:
				addOrder(a.Floor, down, hall)
			case a.Button == elevio.BT_Cab:
				addOrder(a.Floor, 0, cab)
			}

			// fmt.Printf("Added order, length of elevatorOrders is now: %d\n", len(elevatorOrders))
			// fmt.Printf("Added order, elevatorOrders is now: %v\n", elevatorOrders)
			fmt.Printf("Added order, current direction is now: %v\n", d)
			// fmt.Printf("Added order, positionArray is now: %v\n", posArray)

			sortAllOrders(&elevatorOrders, d, posArray)
			// fmt.Printf("Sorted order, length of elevatorOrders is now: %d\n", len(elevatorOrders))

			first_element := elevatorOrders[0]

			// fmt.Printf("Sorted order\n")

			fmt.Printf("ElevatorOrders is now: %v\n", elevatorOrders)

			// Sending the first element of elevatorOrders through the drv_newOrder channel
			// We don't have to worry about the possibility of it being the same order that we are attending to
			// This is because we only set the current direction to the same as it was
			unlockMutexes(&mutex_elevatorOrders, &mutex_d, &mutex_posArray)
			drv_newOrder <- first_element
		}
	}
}
