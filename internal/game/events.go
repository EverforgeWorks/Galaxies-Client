package game

import (
	"fmt"
	"math/rand"
)

// ProcessArrivalEvents calculates and applies random incidents based on RNG.
// It directly mutates the passed Ship struct and returns a log of what happened.
func ProcessArrivalEvents(ship *Ship) []TravelEvent {
	var events []TravelEvent

	// 1. FUEL LEAK CHECK (10% Chance)
	// Mechanical failures are common in the prototyping phase.
	if rand.Float64() < 0.10 {
		// Lose between 5% and 15% of current fuel
		lossPct := 0.05 + rand.Float64()*0.10
		lossAmount := int64(float64(ship.Fuel) * lossPct)

		if lossAmount > 0 {
			ship.Fuel -= lossAmount
			events = append(events, TravelEvent{
				Type:        "fuel_leak",
				Description: "Micrometeoroid impact on fuel line! Emergency seal deployed.",
				Effect:      fmt.Sprintf("Lost %d Fuel", lossAmount),
			})
		}
	}

	// 2. CARGO LOSS CHECK (5% Chance)
	// Requires active cargo contracts.
	if rand.Float64() < 0.05 && len(ship.ActiveContracts) > 0 {
		// Filter for cargo contracts
		var cargoIndices []int
		for i, c := range ship.ActiveContracts {
			if c.Type == "cargo" {
				cargoIndices = append(cargoIndices, i)
			}
		}

		if len(cargoIndices) > 0 {
			// Pick a random contract to fail
			targetIdx := cargoIndices[rand.Intn(len(cargoIndices))]
			lostContract := ship.ActiveContracts[targetIdx]

			// Remove it from the ship (Slicing trick)
			ship.ActiveContracts = append(
				ship.ActiveContracts[:targetIdx],
				ship.ActiveContracts[targetIdx+1:]...,
			)

			events = append(events, TravelEvent{
				Type:        "cargo_loss",
				Description: fmt.Sprintf("Containment breach! %s jettisoned to prevent hull damage.", lostContract.ItemName),
				Effect:      "Contract Failed (Item Lost)",
			})
		}
	}

	// 3. PASSENGER INCIDENT CHECK (5% Chance)
	// Requires active passenger contracts.
	if rand.Float64() < 0.05 && len(ship.ActiveContracts) > 0 {
		// Filter for passenger contracts
		var paxIndices []int
		for i, c := range ship.ActiveContracts {
			if c.Type == "passenger" {
				paxIndices = append(paxIndices, i)
			}
		}

		if len(paxIndices) > 0 {
			targetIdx := paxIndices[rand.Intn(len(paxIndices))]
			// Remove it
			ship.ActiveContracts = append(
				ship.ActiveContracts[:targetIdx],
				ship.ActiveContracts[targetIdx+1:]...,
			)

			events = append(events, TravelEvent{
				Type:        "passenger_loss",
				Description: "Passenger demanded emergency disembark at a waypoint station.",
				Effect:      "Contract Voided",
			})
		}
	}

	return events
}
