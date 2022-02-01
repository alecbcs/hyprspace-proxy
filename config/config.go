package config

type Config struct {
	General general
	Proxy   Proxy
	Clients []Client
}

type general struct {
	Version string
}

type Proxy struct {
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
			Version: "0.2.0",
		},
		Proxy: envParseProxy(Proxy{
			// Default values for the internal proxy config.
			// May be overwritten by environment variables if present.
			Port:    "8001",
			Address: "10.1.1.1/24",
		}),
		Clients: envParseClients(),
	}
}
