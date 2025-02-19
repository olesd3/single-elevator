package main

import (
	"Driver-go/elevio"
	"fmt"
	"sync"
)

const numFloors = 4

var (
	elevatorOrders []Order
	mu1            sync.Mutex
)

var (
	posArray [2*numFloors - 1]bool
	mu2      sync.Mutex
)

var (
	lastFloor int
	mu3       sync.Mutex
)

var (
	currentOrder Order
	mu4          sync.Mutex
)

/*
var (
	lastDir MotorDirection
	mu5     sync.Mutex
)

var (
	drv_dir = make(chan elevio.MotorDirection)
	mu6     sync.Mutex
)
*/

func main() {

	elevio.Init("localhost:15657", numFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_floors2 := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	drv_idle := make(chan bool)
	drv_newOrderTrigger := make(chan bool)

	go elevio.PollButtons(drv_buttons)         // Starts checking for button updates
	go elevio.PollFloorSensor(drv_floors)      // Starts checking for floors updates
	go elevio.PollFloorSensor2(drv_floors2)    // Starts checking for floors updates (for tracking position)
	go elevio.PollObstructionSwitch(drv_obstr) // Starts checking for obstruction updates
	go elevio.PollStopButton(drv_stop)         // Starts checking for stop button presses
	go handleOrders(drv_newOrderTrigger, )

	mu3.Lock()
	lastFloor = 0
	mu3.Unlock()

	mu4.Lock()
	currentOrder = Order{floor: -1, direction: up, orderType: cab}
	mu4.Unlock()

	mu5.Lock()
	lastDir = MD_Stop
	mu5.Unlock()

	var d elevio.MotorDirection = elevio.MD_Down
	elevio.SetMotorDirection(d)
	init := true

	go elevio.GlobalIdleUpdate(drv_idle, &d) // Starts checking for idle updates

	go trackPosition(drv_floors2, &d) // Starts tracking the position of the elevator

	for {
		select {
		case a := <-drv_buttons:
			if !init {

				// Turn on button lamp
				elevio.SetButtonLamp(a.Button, a.Floor, true)

				// Add order
				if a.Button == elevio.BT_HallUp { // If its a hall button
					mu1.Lock()
					addOrder(a.Floor, up, hall)
					mu1.Unlock()
				} else if a.Button == elevio.BT_HallDown { // If its a cab button
					mu1.Lock()
					addOrder(a.Floor, down, hall)
					mu1.Unlock()
				} else if a.Button == elevio.BT_Cab { // If its a cab button
					mu1.Lock()
					addOrder(a.Floor, 0, cab)
					mu1.Unlock()
				}

				fmt.Printf("Elevator orders: %+v\n", elevatorOrders)

				// Handle orders
				

				fmt.Printf("Elevator orders after sorting: %+v\n", elevatorOrders)

			}

		case a := <-drv_floors:
			if a != -1 {

				// Decimal floor
				if a > lastFloor {
					mu5.Lock()
					lastDir = MD_Up
					mu5.Unlock()
				} else if a < lastFloor {
					mu5.Lock()
					lastDir = MD_Down
					mu5.Unlock()
				}

				mu3.Lock()
				lastFloor = a
				mu3.Unlock()

				fmt.Printf("Current floor: %+v | Last direction: %+v\n", a, lastDir)
			}

			if init {
				if a == 0 {
					d = elevio.MD_Stop
					elevio.SetMotorDirection(d)

					init = false

					mu3.Lock()
					lastFloor = a
					mu3.Unlock()
				} else {
					break
				}
			} else {
				fmt.Printf("Current floor: %+v\n", a)
				if a == numFloors-1 {
					d = elevio.MD_Down
				} else if a == 0 {
					d = elevio.MD_Up
				}
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_obstr:
			if !init {
				fmt.Printf("Obstruction update: %+v\n", a)
				if a {
					elevio.SetMotorDirection(elevio.MD_Stop)
				} else {
					elevio.SetMotorDirection(d)
				}
			}

		case a := <-drv_stop:
			fmt.Printf("Stop button: %+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
