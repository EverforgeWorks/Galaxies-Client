package main

import (
	"context"
	"fmt"
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

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load the Universe configuration
	if err := game.LoadConfig(); err != nil {
		log.Printf("CRITICAL: Failed to load universe config: %v", err)
	}

	// Start the Economy Heartbeat
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-ticker.C:
				updatedPlanets := game.ReplenishMarket()
				if len(updatedPlanets) > 0 {
					runtime.EventsEmit(a.ctx, "market_pulse", updatedPlanets)
				}
			}
		}
	}()
}

// -----------------------------------------------------------------------------
// PERSISTENCE & SESSION METHODS
// -----------------------------------------------------------------------------

// NewGameParams defines the data needed to start a fresh journey
type NewGameParams struct {
	Slot        int    `json:"slot"`
	PlayerName  string `json:"player_name"`
	ShipName    string `json:"ship_name"`
	ShipTypeKey string `json:"ship_type_key"`
}

// CreateNewGame initializes a new player and ship based on onboarding choices
func (a *App) CreateNewGame(params NewGameParams) string {
	// Initialize the market for a fresh start
	game.ReplenishMarket()

	// Logic to initialize game.CurrentPlayer and game.CurrentPlayer.ActiveShip
	// based on params. This assumes internal/game has an Init function.
	err := game.InitializeNewPlayer(params.PlayerName, params.ShipName, params.ShipTypeKey)
	if err != nil {
		return "CREATION FAILED: " + err.Error()
	}

	// Save immediately to the chosen slot
	return a.SaveGame(params.Slot)
}

// SaveGame triggers a save to a specific slot file
func (a *App) SaveGame(slot int) string {
	filename := fmt.Sprintf("save_slot_%d.yaml", slot)
	err := game.SaveGame(filename)
	if err != nil {
		return "SAVE FAILED: " + err.Error()
	}
	return "GAME SAVED"
}

// LoadGame triggers a load from a specific slot file
func (a *App) LoadGame(slot int) string {
	filename := fmt.Sprintf("save_slot_%d.yaml", slot)
	err := game.LoadGame(filename)
	if err != nil {
		return "LOAD FAILED: " + err.Error()
	}

	// Initial Market Seed for the loaded session
	game.ReplenishMarket()

	runtime.EventsEmit(a.ctx, "market_pulse", []string{"LOADED"})
	return "GAME LOADED"
}

// -----------------------------------------------------------------------------
// HELPER METHODS
// -----------------------------------------------------------------------------

func getActiveShip() *game.Ship {
	key := game.CurrentPlayer.ActiveShipKey
	return game.CurrentPlayer.Ships[key]
}

func (a *App) enrichShipData(s *game.Ship) *game.Ship {
	s.TotalMass = game.CalculateTotalMass(s)
	s.CurrentBurn = game.CalculateCurrentBurn(s)
	return s
}

// -----------------------------------------------------------------------------
// SHIP & NAVIGATION METHODS
// -----------------------------------------------------------------------------

type PlayerStateResponse struct {
	PlayerName string     `json:"player_name"`
	Credits    int        `json:"credits"`
	Ship       *game.Ship `json:"ship"`
}

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

func (a *App) GetPlanets() []game.Planet {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()
	return game.CurrentUniverse.Planets
}

func (a *App) Travel(destinationKey string) TravelResponse {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()
	dest := game.GetPlanet(destinationKey)
	curr := game.GetPlanet(ship.LocationKey)

	if dest == nil {
		return TravelResponse{Success: false, Error: "Invalid Destination"}
	}

	dist := game.CalculateDistance(curr.Coordinates, dest.Coordinates)
	burn := game.CalculateCurrentBurn(ship)
	cost := dist * burn

	if ship.Fuel < cost {
		return TravelResponse{Success: false, Error: "Insufficient Fuel"}
	}

	ship.Fuel -= cost
	ship.LocationKey = dest.Key

	events := game.ProcessArrivalEvents(ship)

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
	game.CurrentPlayer.Credits += payout

	return TravelResponse{
		Success: true,
		State: PlayerStateResponse{
			PlayerName: game.CurrentPlayer.Name,
			Credits:    game.CurrentPlayer.Credits,
			Ship:       a.enrichShipData(ship),
		},
		Events:   events,
		Duration: dist,
	}
}

type TravelQuoteResponse struct {
	Distance          int64 `json:"distance"`
	FuelCost          int64 `json:"fuel_cost"`
	CanAfford         bool  `json:"can_afford"`
	BurnRate          int64 `json:"burn_rate"`
	EstimatedDuration int64 `json:"estimated_duration_seconds"`
}

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

func (a *App) GetAvailableContracts() []game.Contract {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()
	return game.AvailableContracts[ship.LocationKey]
}

func (a *App) AcceptJob(contractID string) bool {
	game.DataLock.Lock()
	defer game.DataLock.Unlock()

	ship := getActiveShip()
	loc := ship.LocationKey
	board := game.AvailableContracts[loc]

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

	ship.ActiveContracts = append(ship.ActiveContracts, target)
	game.AvailableContracts[loc] = append(board[:idx], board[idx+1:]...)

	game.Market.RecordAcceptance(target.OriginKey, target.ItemKey, target.Quantity)

	return true
}

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

func (a *App) GetModules() []game.ShipModule {
	game.DataLock.RLock()
	defer game.DataLock.RUnlock()

	ship := getActiveShip()
	if ship.LocationKey != "planet_prime" {
		return []game.ShipModule{}
	}
	return game.CurrentUniverse.ShipModules
}

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
