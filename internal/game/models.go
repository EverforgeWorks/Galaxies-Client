package game

type GameBalance struct {
	StartingCredits    int `yaml:"starting_credits" json:"starting_credits"`
	FuelCostPerUnit    int `yaml:"fuel_cost_per_unit" json:"fuel_cost_per_unit"`
	FuelMassPerUnit    int `yaml:"fuel_mass_per_unit" json:"fuel_mass_per_unit"`
	DistancePayoutMult int `yaml:"distance_payout_mult" json:"distance_payout_mult"`
}

type ShipModule struct {
	Key          string `yaml:"key" json:"key"`
	Name         string `yaml:"name" json:"name"`
	Description  string `yaml:"description" json:"description"`
	Cost         int    `yaml:"cost" json:"cost"`
	StatModifier string `yaml:"stat_modifier" json:"stat_modifier"`
	StatValue    int    `yaml:"stat_value" json:"stat_value"`
}

type Contract struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	ItemName       string `json:"item_name"`
	ItemKey        string `json:"item_key"`
	Quantity       int    `json:"quantity"`
	MassPerUnit    int    `json:"mass_per_unit"`
	OriginKey      string `json:"origin_key"`
	DestinationKey string `json:"destination_key"`
	Payout         int    `json:"payout"`
}

type Planet struct {
	Key           string   `json:"key" yaml:"key"`
	Name          string   `json:"name" yaml:"name"`
	Coordinates   []int    `json:"coordinates" yaml:"coordinates"`
	Production    []string `json:"production" yaml:"production"`
	Demand        []string `json:"demand" yaml:"demand"`
	MinCargo      int      `json:"min_cargo" yaml:"min_cargo"`
	MaxCargo      int      `json:"max_cargo" yaml:"max_cargo"`
	MinPassengers int      `json:"min_passengers" yaml:"min_passengers"`
	MaxPassengers int      `json:"max_passengers" yaml:"max_passengers"`
}

// Player represents the human user.
type Player struct {
	Name          string           `json:"name"`
	Credits       int              `json:"credits"`
	Ships         map[string]*Ship `json:"ships"`           // Map of [InstanceID] -> Ship
	ActiveShipKey string           `json:"active_ship_key"` // The ID of the ship currently being flown
}

// ShipTemplate defines the base stats for a model of ship (loaded from YAML).
type ShipTemplate struct {
	Key            string `yaml:"key"`
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	MaxFuel        int64  `yaml:"max_fuel"`
	BaseBurnRate   int64  `yaml:"base_burn_rate"`
	BurnDamping    int64  `yaml:"burn_damping"`
	BaseMass       int64  `yaml:"base_mass"`
	CargoCapacity  int    `yaml:"cargo_capacity"`
	PassengerSlots int    `yaml:"passenger_slots"`
	MaxModuleSlots int    `yaml:"max_module_slots"`
}

// Ship represents a specific instance of a vessel owned by a player.
type Ship struct {
	InstanceID  string `json:"instance_id"`
	TemplateKey string `json:"template_key"`
	Name        string `json:"name"`
	LocationKey string `json:"location_key"`
	Fuel        int64  `json:"fuel"`
	MaxFuel     int64  `json:"max_fuel"`

	BaseBurnRate int64 `json:"base_burn_rate"`
	BurnDamping  int64 `json:"burn_damping"`
	BaseMass     int64 `json:"base_mass"`

	CargoCapacity  int `json:"cargo_capacity"`
	PassengerSlots int `json:"passenger_slots"`
	MaxModuleSlots int `json:"max_module_slots"`

	InstalledModules []ShipModule `json:"installed_modules"`
	ActiveContracts  []Contract   `json:"active_contracts"`

	// Dynamic Fields
	TotalMass   int64 `json:"total_mass" yaml:"-"`
	CurrentBurn int64 `json:"current_burn" yaml:"-"`
}

type TravelEvent struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Effect      string `json:"effect"`
}

type PassengerConfig struct {
	BaseTicketPrice  int `yaml:"base_ticket_price"`
	MassPerPassenger int `yaml:"mass_per_passenger"`
}

type Commodity struct {
	Key       string `yaml:"key" json:"key"`
	Name      string `yaml:"name" json:"name"`
	BaseValue int    `yaml:"base_value" json:"base_value"`
	Mass      int    `yaml:"mass" json:"mass"`
}

type Universe struct {
	BalanceConfig   GameBalance     `yaml:"game_balance"`
	ShipTemplates   []ShipTemplate  `yaml:"ship_templates"` // Changed from PlayerShipConfig
	Commodities     []Commodity     `yaml:"commodities"`
	Planets         []Planet        `yaml:"planets"`
	ShipModules     []ShipModule    `yaml:"ship_modules"`
	PassengerConfig PassengerConfig `yaml:"passenger_config"`
}

type MarketState struct {
	SourceHeat map[string]map[string]float64
	DestHeat   map[string]map[string]float64
}

// SaveData container for persistence
type SaveData struct {
	Player    Player                `yaml:"player"`
	Market    MarketState           `yaml:"market"`
	Contracts map[string][]Contract `yaml:"contracts"`
}
