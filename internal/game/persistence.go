/*
Package game
File: persistence.go
Description:
    Handles saving and loading the game state to/from a YAML file.
    It serializes the SaveData struct defined in models.go.
*/

package game

import (
	"os"

	"gopkg.in/yaml.v3"
)

// SaveGame writes the current state to a YAML file.
func SaveGame(filename string) error {
	DataLock.RLock() // Read Lock (we are reading state to save it)
	defer DataLock.RUnlock()

	// 1. Pack the state into the SaveData container
	data := SaveData{
		Player:    CurrentPlayer,
		Market:    Market,
		Contracts: AvailableContracts,
	}

	// 2. Marshal to YAML
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// 3. Write File to disk
	// 0644 = User R/W, Group R, World R
	return os.WriteFile(filename, bytes, 0644)
}

// LoadGame reads a YAML file and overwrites the memory state.
func LoadGame(filename string) error {
	DataLock.Lock() // Write Lock (we are overwriting the entire state)
	defer DataLock.Unlock()

	// 1. Read File
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// 2. Unmarshal into a temp struct
	var data SaveData
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return err
	}

	// 3. Restore Global State
	CurrentPlayer = data.Player
	Market = data.Market
	AvailableContracts = data.Contracts

	// Note:
	// Fields tagged with `yaml:"-"` (TotalMass, CurrentBurn) are not saved.
	// They will be recalculated automatically the next time 'enrichShipData'
	// is called in app.go, so we don't need to manually re-compute them here.

	return nil
}
