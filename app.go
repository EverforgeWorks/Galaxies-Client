package main

import (
	"context"
	"log"
	"time"

	"galaxies-client/internal/game" // Import local game logic

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Load the Universe (Now from local file in the client root)
	if err := game.LoadConfig(); err != nil {
		log.Printf("CRITICAL: Failed to load universe config: %v", err)
	}

	// 2. Initial Market Seed
	game.ReplenishMarket()

	// 3. Start the Economy Heartbeat (Background Simulation)
	// This replaces the server's main.go loop
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-a.ctx.Done():
				return // Stop when app closes
			case <-ticker.C:
				// Run logic
				updatedPlanets := game.ReplenishMarket()

				if len(updatedPlanets) > 0 {
					// Emit event to Frontend (Replaces Websocket Broadcast)
					runtime.EventsEmit(a.ctx, "market_pulse", updatedPlanets)
					log.Printf("SIMULATION: Updated %d planets", len(updatedPlanets))
				}
			}
		}
	}()
}

// -----------------------------------------------------------------------------
// SHIP & NAVIGATION METHODS
// -----------------------------------------------------------------------------

// GetShipState returns the player's current status
func (a *App) GetShipState() game.Ship {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()
	return game.PlayerShip
}

// GetPlanets returns the static universe map
func (a *App) GetPlanets() []game.Planet {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()
	return game.CurrentUniverse.Planets
}

// TravelResponse is the struct returned to the UI after a travel action
type TravelResponse struct {
	Success  bool               `json:"success"`
	Ship     game.Ship          `json:"ship"`
	Events   []game.TravelEvent `json:"events"`
	Duration int64              `json:"duration_seconds"`
	Error    string             `json:"error,omitempty"`
}

// Travel executes a move to another planet
func (a *App) Travel(destinationKey string) TravelResponse {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	dest := game.GetPlanet(destinationKey)
	curr := game.GetPlanet(game.PlayerShip.LocationKey)

	if dest == nil {
		return TravelResponse{Success: false, Error: "Invalid Destination"}
	}

	// Physics
	dist := game.CalculateDistance(curr.Coordinates, dest.Coordinates)
	burn := game.CalculateCurrentBurn()
	cost := dist * burn

	if game.PlayerShip.Fuel < cost {
		return TravelResponse{Success: false, Error: "Insufficient Fuel"}
	}

	// Execute Move
	game.PlayerShip.Fuel -= cost
	game.PlayerShip.LocationKey = dest.Key

	// Events
	events := game.ProcessArrivalEvents(&game.PlayerShip)

	// Deliveries
	payout := 0
	remaining := []game.Contract{}
	for _, c := range game.PlayerShip.ActiveContracts {
		if c.DestinationKey == dest.Key {
			payout += c.Payout
			game.Market.RecordDelivery(c.DestinationKey, c.ItemKey, c.Quantity)
		} else {
			remaining = append(remaining, c)
		}
	}
	game.PlayerShip.ActiveContracts = remaining
	game.PlayerShip.Credits += payout

	return TravelResponse{
		Success:  true,
		Ship:     game.PlayerShip,
		Events:   events,
		Duration: dist, // 1 Second per LY
	}
}

type TravelQuoteResponse struct {
	Distance          int64 `json:"distance"`
	FuelCost          int64 `json:"fuel_cost"`
	CanAfford         bool  `json:"can_afford"`
	BurnRate          int64 `json:"burn_rate"`
	EstimatedDuration int64 `json:"estimated_duration_seconds"`
}

// GetTravelQuote calculates cost without moving
func (a *App) GetTravelQuote(destinationKey string) TravelQuoteResponse {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	dest := game.GetPlanet(destinationKey)
	curr := game.GetPlanet(game.PlayerShip.LocationKey)

	if dest == nil {
		return TravelQuoteResponse{}
	}

	dist := game.CalculateDistance(curr.Coordinates, dest.Coordinates)
	burn := game.CalculateCurrentBurn()
	cost := dist * burn

	return TravelQuoteResponse{
		Distance:          dist,
		FuelCost:          cost,
		CanAfford:         game.PlayerShip.Fuel >= cost,
		BurnRate:          burn,
		EstimatedDuration: dist,
	}
}

// Refuel fills the tank
func (a *App) Refuel() bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	needed := game.PlayerShip.MaxFuel - game.PlayerShip.Fuel
	if needed <= 0 {
		return false
	}

	cost := (int(needed) / 100) * game.CurrentUniverse.BalanceConfig.FuelCostPerUnit
	if game.PlayerShip.Credits < cost {
		return false
	}

	game.PlayerShip.Credits -= cost
	game.PlayerShip.Fuel = game.PlayerShip.MaxFuel
	return true
}

// -----------------------------------------------------------------------------
// ECONOMY & CONTRACT METHODS
// -----------------------------------------------------------------------------

// GetAvailableContracts returns jobs at the current location
func (a *App) GetAvailableContracts() []game.Contract {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	loc := game.PlayerShip.LocationKey
	return game.AvailableContracts[loc]
}

// AcceptJob moves a contract from planet to ship
func (a *App) AcceptJob(contractID string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	loc := game.PlayerShip.LocationKey
	board := game.AvailableContracts[loc]

	// Find contract
	idx := -1
	var target game.Contract
	for i, c := range board {
		if c.ID == contractID {
			idx = i
			target = c
			break
		}
	}

	if idx == -1 {
		return false
	}

	// Validate Capacity
	currentCargo, currentPax := 0, 0
	for _, c := range game.PlayerShip.ActiveContracts {
		if c.Type == "cargo" {
			currentCargo += c.Quantity
		} else {
			currentPax += c.Quantity
		}
	}

	if target.Type == "cargo" && currentCargo+target.Quantity > game.PlayerShip.CargoCapacity {
		return false
	}
	if target.Type == "passenger" && currentPax+target.Quantity > game.PlayerShip.PassengerSlots {
		return false
	}

	// Move Contract
	game.PlayerShip.ActiveContracts = append(game.PlayerShip.ActiveContracts, target)
	game.AvailableContracts[loc] = append(board[:idx], board[idx+1:]...)

	// Economy Impact
	game.Market.RecordAcceptance(target.OriginKey, target.ItemKey, target.Quantity)

	return true
}

// DropJob abandons a contract
func (a *App) DropJob(contractID string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	idx := -1
	for i, c := range game.PlayerShip.ActiveContracts {
		if c.ID == contractID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	game.PlayerShip.ActiveContracts = append(
		game.PlayerShip.ActiveContracts[:idx],
		game.PlayerShip.ActiveContracts[idx+1:]...,
	)
	return true
}

// -----------------------------------------------------------------------------
// MODULE & UPGRADE METHODS
// -----------------------------------------------------------------------------

// GetModules returns upgrades (Only at Prime)
func (a *App) GetModules() []game.ShipModule {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	if game.PlayerShip.LocationKey != "planet_prime" {
		return []game.ShipModule{}
	}
	return game.CurrentUniverse.ShipModules
}

// BuyModule purchases an upgrade
func (a *App) BuyModule(key string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	if game.PlayerShip.LocationKey != "planet_prime" {
		return false
	}

	// Check Slots
	if len(game.PlayerShip.InstalledModules) >= game.PlayerShip.MaxModuleSlots {
		return false
	}

	mod := game.GetModule(key)
	if mod == nil || game.PlayerShip.Credits < mod.Cost {
		return false
	}

	// Apply
	game.PlayerShip.Credits -= mod.Cost
	game.PlayerShip.InstalledModules = append(game.PlayerShip.InstalledModules, *mod)

	// Update Stats
	switch mod.StatModifier {
	case "cargo_capacity":
		game.PlayerShip.CargoCapacity += mod.StatValue
	case "passenger_slots":
		game.PlayerShip.PassengerSlots += mod.StatValue
	case "max_fuel":
		game.PlayerShip.MaxFuel += int64(mod.StatValue)
	case "base_burn_rate":
		game.PlayerShip.BaseBurnRate += int64(mod.StatValue)
	}

	return true
}
