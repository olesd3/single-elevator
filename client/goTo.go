package main

import (
	"Driver-go/elevio"
)

func goToFloor(floor int, d *elevio.MotorDirection) {
	for floor != elevio.GetFloor() {
		if elevio.GetFloor() != -1 {
			if floor > elevio.GetFloor() {
				*d = elevio.MotorDirection(elevio.MD_Up)
				elevio.SetMotorDirection(*d)
			} else if floor < elevio.GetFloor() {
				*d = elevio.MotorDirection(elevio.MD_Down)
				elevio.SetMotorDirection(*d)
			}
		}
	}

	*d = elevio.MotorDirection(elevio.MD_Stop)
	elevio.SetMotorDirection(*d)
}
