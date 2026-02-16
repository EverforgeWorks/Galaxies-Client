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
// HELPER METHODS
// -----------------------------------------------------------------------------

// getActiveShip retrieves the pointer to the ship currently being flown by the player.
// Assumes the caller holds the DataLock.
func getActiveShip() *game.Ship {
	key := game.CurrentPlayer.ActiveShipKey
	return game.CurrentPlayer.Ships[key]
}

// enrichShipData calculates dynamic physics properties (Mass/Burn) for the ship
// before sending it to the UI.
func (a *App) enrichShipData(s *game.Ship) *game.Ship {
	s.TotalMass = game.CalculateTotalMass(s)
	s.CurrentBurn = game.CalculateCurrentBurn(s)
	return s
}

// -----------------------------------------------------------------------------
// SHIP & NAVIGATION METHODS
// -----------------------------------------------------------------------------

// PlayerStateResponse combines Player info with the Active Ship info.
type PlayerStateResponse struct {
	PlayerName string     `json:"player_name"`
	Credits    int        `json:"credits"`
	Ship       *game.Ship `json:"ship"`
}

// GetShipState returns the Player + Active Ship status.
func (a *App) GetShipState() PlayerStateResponse {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()

	return PlayerStateResponse{
		PlayerName: game.CurrentPlayer.Name,
		Credits:    game.CurrentPlayer.Credits,
		Ship:       a.enrichShipData(ship),
	}
}

// GetPlanets returns the static universe map.
func (a *App) GetPlanets() []game.Planet {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()
	return game.CurrentUniverse.Planets
}

// TravelResponse is the struct returned to the UI after a travel action.
type TravelResponse struct {
	Success  bool                `json:"success"`
	State    PlayerStateResponse `json:"state"` // Returns full state update
	Events   []game.TravelEvent  `json:"events"`
	Duration int64               `json:"duration_seconds"`
	Error    string              `json:"error,omitempty"`
}

// Travel executes a move to another planet using the Active Ship.
func (a *App) Travel(destinationKey string) TravelResponse {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()
	dest := game.GetPlanet(destinationKey)
	curr := game.GetPlanet(ship.LocationKey)

	if dest == nil {
		return TravelResponse{Success: false, Error: "Invalid Destination"}
	}

	// Physics on specific ship instance
	dist := game.CalculateDistance(curr.Coordinates, dest.Coordinates)
	burn := game.CalculateCurrentBurn(ship)
	cost := dist * burn

	if ship.Fuel < cost {
		return TravelResponse{Success: false, Error: "Insufficient Fuel"}
	}

	// Execute Move
	ship.Fuel -= cost
	ship.LocationKey = dest.Key

	// Events
	events := game.ProcessArrivalEvents(ship)

	// Deliveries
	payout := 0
	remaining := []game.Contract{}
	for _, c := range ship.ActiveContracts {
		if c.DestinationKey == dest.Key {
			payout += c.Payout
			game.Market.RecordDelivery(c.DestinationKey, c.ItemKey, c.Quantity)
		} else {
			remaining = append(remaining, c)
		}
	}
	ship.ActiveContracts = remaining
	game.CurrentPlayer.Credits += payout // Credits go to Player Wallet

	return TravelResponse{
		Success: true,
		State: PlayerStateResponse{
			PlayerName: game.CurrentPlayer.Name,
			Credits:    game.CurrentPlayer.Credits,
			Ship:       a.enrichShipData(ship),
		},
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

// GetTravelQuote calculates cost without moving.
func (a *App) GetTravelQuote(destinationKey string) TravelQuoteResponse {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()
	dest := game.GetPlanet(destinationKey)
	curr := game.GetPlanet(ship.LocationKey)

	if dest == nil {
		return TravelQuoteResponse{}
	}

	dist := game.CalculateDistance(curr.Coordinates, dest.Coordinates)
	burn := game.CalculateCurrentBurn(ship)
	cost := dist * burn

	return TravelQuoteResponse{
		Distance:          dist,
		FuelCost:          cost,
		CanAfford:         ship.Fuel >= cost,
		BurnRate:          burn,
		EstimatedDuration: dist,
	}
}

// Refuel fills the tank of the Active Ship using Player Credits.
func (a *App) Refuel() bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()
	needed := ship.MaxFuel - ship.Fuel
	if needed <= 0 {
		return false
	}

	cost := (int(needed) / 100) * game.CurrentUniverse.BalanceConfig.FuelCostPerUnit
	if game.CurrentPlayer.Credits < cost {
		return false
	}

	game.CurrentPlayer.Credits -= cost
	ship.Fuel = ship.MaxFuel
	return true
}

// -----------------------------------------------------------------------------
// ECONOMY & CONTRACT METHODS
// -----------------------------------------------------------------------------

// GetAvailableContracts returns jobs at the current location of the Active Ship.
func (a *App) GetAvailableContracts() []game.Contract {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()
	return game.AvailableContracts[ship.LocationKey]
}

// AcceptJob moves a contract from planet to the Active Ship.
func (a *App) AcceptJob(contractID string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()
	loc := ship.LocationKey
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
	for _, c := range ship.ActiveContracts {
		if c.Type == "cargo" {
			currentCargo += c.Quantity
		} else {
			currentPax += c.Quantity
		}
	}

	if target.Type == "cargo" && currentCargo+target.Quantity > ship.CargoCapacity {
		return false
	}
	if target.Type == "passenger" && currentPax+target.Quantity > ship.PassengerSlots {
		return false
	}

	// Move Contract
	ship.ActiveContracts = append(ship.ActiveContracts, target)
	game.AvailableContracts[loc] = append(board[:idx], board[idx+1:]...)

	// Economy Impact
	game.Market.RecordAcceptance(target.OriginKey, target.ItemKey, target.Quantity)

	return true
}

// DropJob abandons a contract from the Active Ship.
func (a *App) DropJob(contractID string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()

	idx := -1
	for i, c := range ship.ActiveContracts {
		if c.ID == contractID {
			idx = i
			break
		}
	}

	if idx == -1 {
		return false
	}

	ship.ActiveContracts = append(
		ship.ActiveContracts[:idx],
		ship.ActiveContracts[idx+1:]...,
	)
	return true
}

// -----------------------------------------------------------------------------
// MODULE & UPGRADE METHODS
// -----------------------------------------------------------------------------

// GetModules returns upgrades (Only at Prime).
func (a *App) GetModules() []game.ShipModule {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()
	if ship.LocationKey != "planet_prime" {
		return []game.ShipModule{}
	}
	return game.CurrentUniverse.ShipModules
}

// BuyModule purchases an upgrade for the Active Ship.
func (a *App) BuyModule(key string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()

	if ship.LocationKey != "planet_prime" {
		return false
	}
	if len(ship.InstalledModules) >= ship.MaxModuleSlots {
		return false
	}

	mod := game.GetModule(key)
	if mod == nil || game.CurrentPlayer.Credits < mod.Cost {
		return false
	}

	game.CurrentPlayer.Credits -= mod.Cost
	ship.InstalledModules = append(ship.InstalledModules, *mod)

	// Apply Stats to SHIP
	switch mod.StatModifier {
	case "cargo_capacity":
		ship.CargoCapacity += mod.StatValue
	case "passenger_slots":
		ship.PassengerSlots += mod.StatValue
	case "max_fuel":
		ship.MaxFuel += int64(mod.StatValue)
	case "base_burn_rate":
		ship.BaseBurnRate += int64(mod.StatValue)
	}

	return true
}

// -----------------------------------------------------------------------------
// PERSISTENCE METHODS
// -----------------------------------------------------------------------------

// SaveGame triggers a save to 'savegame.yaml' in the game directory.
func (a *App) SaveGame() string {
	err := game.SaveGame("savegame.yaml")
	if err != nil {
		return "SAVE FAILED: " + err.Error()
	}
	return "GAME SAVED"
}

// LoadGame triggers a load from 'savegame.yaml'.
func (a *App) LoadGame() string {
	err := game.LoadGame("savegame.yaml")
	if err != nil {
		return "LOAD FAILED: " + err.Error()
	}

	// Force a UI refresh event immediately after loading
	runtime.EventsEmit(a.ctx, "market_pulse", []string{"LOADED"})

	return "GAME LOADED"
}
