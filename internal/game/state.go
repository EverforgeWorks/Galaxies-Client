package game

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Global State
var (
	CurrentUniverse    Universe
	CurrentPlayer      Player // Replaces PlayerShip
	AvailableContracts map[string][]Contract
	Market             MarketState
	DataLock           sync.RWMutex
)

// LoadConfig reads universe.yaml and initializes the game state.
func LoadConfig() error {
	DataLock.Lock()
	defer DataLock.Unlock()

	// 1. Read YAML
	data, err := os.ReadFile("universe.yaml")
	if err != nil {
		return err
	}

	// 2. Parse Universe
	if err := yaml.Unmarshal(data, &CurrentUniverse); err != nil {
		return err
	}

	// 3. Initialize Runtime State (New Game)
	// Create the default player
	CurrentPlayer = Player{
		Name:          "Cmdr. Haddock",
		Credits:       CurrentUniverse.BalanceConfig.StartingCredits,
		Ships:         make(map[string]*Ship),
		ActiveShipKey: "ship_1",
	}

	// Create Default Ship (Standard Hauler)
	// We find the template in the loaded config
	var starterTemplate ShipTemplate
	for _, t := range CurrentUniverse.ShipTemplates {
		if t.Key == "ship_hauler" { // Default to Hauler
			starterTemplate = t
			break
		}
	}

	// Fallback if config is broken
	if starterTemplate.Key == "" && len(CurrentUniverse.ShipTemplates) > 0 {
		starterTemplate = CurrentUniverse.ShipTemplates[0]
	}

	startingShip := &Ship{
		InstanceID:       "ship_1",
		TemplateKey:      starterTemplate.Key,
		Name:             "SS " + starterTemplate.Name,
		LocationKey:      "planet_prime",
		Fuel:             starterTemplate.MaxFuel,
		MaxFuel:          starterTemplate.MaxFuel,
		BaseBurnRate:     starterTemplate.BaseBurnRate,
		BurnDamping:      starterTemplate.BurnDamping,
		BaseMass:         starterTemplate.BaseMass,
		CargoCapacity:    starterTemplate.CargoCapacity,
		PassengerSlots:   starterTemplate.PassengerSlots,
		MaxModuleSlots:   starterTemplate.MaxModuleSlots,
		InstalledModules: []ShipModule{},
		ActiveContracts:  []Contract{},
	}
	CurrentPlayer.Ships["ship_1"] = startingShip

	// 4. Initialize Market
	Market = MarketState{
		SourceHeat: make(map[string]map[string]float64),
		DestHeat:   make(map[string]map[string]float64),
	}

	// 5. Initialize Job Boards
	AvailableContracts = make(map[string][]Contract)

	return nil
}
