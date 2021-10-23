package config

type Config struct {
	General general
	Relay   Relay
	Clients []Client
}

type general struct {
	Version string
}

type Relay struct {
	ID         string
	Port       string
	Address    string
	PrivateKey string
}

type Client struct {
	ID      string
	Address string
	Ports   []int
}

var Global Config

func init() {
	// Build default config.
	Global = Config{
		General: general{
			Version: "0.0.1",
		},
		Relay: envParseRelay(Relay{
			// Default values for the internal relay config.
			// May be overwritten by environment variables if present.
			Port:    "8001",
			Address: "10.1.1.1/24",
		}),
		Clients: envParseClients(),
	}
}
