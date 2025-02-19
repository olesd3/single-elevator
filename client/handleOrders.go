package main

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// Region: Data types for the orders

type OrderDirection int

const (
	up   OrderDirection = 1
	down OrderDirection = -1
)

type OrderType int

const (
	hall OrderType = 0
	cab  OrderType = 1
)

type Order struct {
	floor     int
	direction OrderDirection // 1 for up, -1 for down
	orderType OrderType      // 0 for hall, 1 for cab
}

// end Region: Data types for the orders

func findHighestOrders(elevatorOrders []Order) []Order {
	// Set the initial highest floor to a very low value
	highestFloor := -1
	var highestOrders []Order

	// Loop through each order to find the highest floor
	for _, order := range elevatorOrders {
		if order.floor > highestFloor {
			// If we find a higher floor, reset the slice and add the current order
			highestFloor = order.floor
			highestOrders = []Order{order}
		} else if order.floor == highestFloor {
			// If the floor matches the current highest, add it to the slice
			highestOrders = append(highestOrders, order)
		}
	}

	return highestOrders
}

func findLowestOrders(elevatorOrders []Order) []Order {
	// Set the initial lowest floor to a very high value
	lowestFloor := 1000000
	var lowestOrders []Order

	// Loop through each order to find the lowest floor
	for _, order := range elevatorOrders {
		if order.floor < lowestFloor {
			// If we find a lower floor, reset the slice and add the current order
			lowestFloor = order.floor
			lowestOrders = []Order{order}
		} else if order.floor == lowestFloor {
			// If the floor matches the current lowest, add it to the slice
			lowestOrders = append(lowestOrders, order)
		}
	}

	return lowestOrders
}

func orderInContainer(order_slice []Order, order_ Order) bool {
	for _, v := range order_slice {
		if v == order_ {
			return true
		}
	}
	return false
}

func sortOrdersInDirection(elevatorOrders []Order, d MotorDirection, posArray [2*numFloors - 1]bool) ([]Order, []Order) {
	//Todo: Lock elevatorOrders with mutex

	highestOrders := findHighestOrders(elevatorOrders)
	lowestOrders := findLowestOrders(elevatorOrders)

	//Calculating the current floor as a decimal so that its compareable to
	currentFloor := float32(0)
	for i := 0; i < 2*numFloors-1; i++ {
		if posArray[i] {
			currentFloor = float32(i) / 2
		}
	}

	//Based current direction => find all the equidirectional orders plus potential extremities
	//Store the relevant orders in relevantOrders and the rest in irrelevantOrders
	relevantOrders := []Order{}
	irrelevantOrders := []Order{}

	for _, order := range elevatorOrders {
		inHighest := orderInContainer(highestOrders, order)
		inLowest := orderInContainer(lowestOrders, order)

		//We define a variable for measuring the distance between current_pos and order.
		//Positive -> The order is above us
		//Zero     -> The order is at the same floor
		//Negative -> The order is below us
		distOrderToCurrent := float32(order.floor) - currentFloor
		switch {
		case (d == MD_Up) && (distOrderToCurrent >= 0.0): //If we're going up and the order is above/same
			switch {
			case inHighest:
				relevantOrders = append(relevantOrders, order)
			case order.direction == up || order.orderType == cab:
				relevantOrders = append(relevantOrders, order)
			case order.direction == down:
				irrelevantOrders = append(irrelevantOrders, order)
			}
		case (d == MD_Up) && (distOrderToCurrent < 0.0): //If we're going up and the order is below/same
			irrelevantOrders = append(irrelevantOrders, order)

		case (d == MD_Down) && (distOrderToCurrent <= 0.0): //If we're going down and the order is below/same
			//If order is down or cab
			switch {
			case inLowest:
				relevantOrders = append(relevantOrders, order)
			case order.direction == down || order.orderType == cab:
				relevantOrders = append(relevantOrders, order)
			case order.direction == up:
				irrelevantOrders = append(irrelevantOrders, order)
			}
		case (d == MD_Down) && (distOrderToCurrent > 0.0): //If we're going down and the order is up/same
			irrelevantOrders = append(irrelevantOrders, order)
		}

	}

	//Now that we've seperated the relevant and irrellevant orders from each other, we sort the relevant part
	//If the current direction is up, we sort them in ascending order
	if d == MD_Up {
		//Perform bubblesort in ascending order
		n := len(relevantOrders)
		for i := 0; i < n-1; i++ {
			// Last i elements are already sorted
			for j := 0; j < n-i-1; j++ {
				if relevantOrders[j].floor > relevantOrders[j+1].floor {
					// Swap arr[j] and arr[j+1]
					relevantOrders[j], relevantOrders[j+1] = relevantOrders[j+1], relevantOrders[j]
				}
			}
		}
	}

	//If the current direction is down, we sort them in descending order
	if d == MD_Down {
		//Perform bubblesort in descending order
		n := len(relevantOrders)
		for i := 0; i < n-1; i++ {
			// Last i elements are already sorted
			for j := 0; j < n-i-1; j++ {
				if relevantOrders[j].floor < relevantOrders[j+1].floor {
					// Swap arr[j] and arr[j+1]
					relevantOrders[j], relevantOrders[j+1] = relevantOrders[j+1], relevantOrders[j]
				}
			}
		}
	}

	return relevantOrders, irrelevantOrders
}

func sortAllOrders(elevatorOrders *[]Order, d MotorDirection, posArray [2*numFloors - 1]bool) {
	if len(*elevatorOrders) == 0 || len(*elevatorOrders) == 1 {
		return
	}

	// Creating the datatypes specfic to our function
	copy_posArray := posArray
	relevantOrders := []Order{}
	irrelevantOrders := []Order{}

	// Start - first section
	firstSection := []Order{}

	irrelevantOrders = *elevatorOrders
	relevantOrders, irrelevantOrders = sortOrdersInDirection(irrelevantOrders, d, copy_posArray)
	firstSection = relevantOrders

	if len(irrelevantOrders) == 0 {
		*elevatorOrders = firstSection
		return
	}
	// End - First Section

	// Start - Second section
	secondSection := []Order{}
	reverseDirection(&d)
	updatePosArray(d, &copy_posArray)

	relevantOrders, irrelevantOrders = sortOrdersInDirection(irrelevantOrders, d, copy_posArray)
	secondSection = relevantOrders

	if len(irrelevantOrders) == 0 {
		*elevatorOrders = append(firstSection, secondSection...)
		return
	}
	// End - Second section

	// Start - Third section
	thirdSection := []Order{}
	reverseDirection(&d)
	updatePosArray(d, &copy_posArray)
	relevantOrders, _ = sortOrdersInDirection(irrelevantOrders, d, copy_posArray)
	thirdSection = relevantOrders
	// End - Third section

	*elevatorOrders = append(firstSection, secondSection...)
	*elevatorOrders = append(*elevatorOrders, thirdSection...)
}
