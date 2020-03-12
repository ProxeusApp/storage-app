package config

import (
	"encoding/json"
	"flag"
	"log"
	"math/big"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// plug pgp-server conf here too, decide how much should be shared across targets, better pkg location

/*
	This configuration can be used in three ways:
	1. Using the defaults specified in flag
	2. Using the specified arguments in flag
	3. Using the config.json in the current dir or as specified by -cfg
*/
type Configuration struct {
	PGPPublicServiceURL string `mapstructure:"pgpPublicServiceURL"`
	EthClientURL        string `mapstructure:"ETHCLIENTURL"`
	EthWebSocketURL     string `mapstructure:"ETHWEBSOCKETURL"`
	MainHostedURL       string `mapstructure:"mainHostedURL"`

	ServiceAddress string `mapstructure:"a"`

	StorageDir             string `mapstructure:"dir"`
	StorageProviderAddress string `mapstructure:"address"`
	ContractAddress        string `mapstructure:"contract"`
	XESContractAddress     string `mapstructure:"xesContract"`
	DevMode                bool   `mapstructure:"devMode"`
	ForceSpp               string `mapstructure:"forceSpp"`
	AutoTLS                bool   `mapstructure:"autotls"`
	TestMode               string `mapstructure:"TESTMODE"`

	PprofDebug bool `mapstructure:"pprof"`

	BlockchainNet string `mapstructure:"blockchainNet"`
}

var Config Configuration

func init() {
	flag.String("pgpPublicServiceURL", "http://localhost:8084", "PGP public service URL")
	flag.String("ethClientURL", "https://ropsten.infura.io/v3/YOURAPIKEY", "Ethereum client URL")
	flag.String("ethWebSocketURL", "wss://ropsten.infura.io/ws/v3/YOURAPIKEY", "Ethereum websocket URL")
	flag.String("mainHostedURL", "https://dev.proxeus.com", "Main hosted URL")

	flag.String("a", ":8082", "pddress and port")

	flag.String("dir", "./", "directory") //0x5b3d62ca34bbef5428f660dffe893f084a985754
	flag.String("address", "0x5C9eDfaaC887552D6b521E38dAA3BFf1f645fD36", "The storage providers ethereum address")
	flag.String("xesContract", "0x84E0b37e8f5B4B86d5d299b0B0e33686405A3919", "XES contract address")
	flag.String("contract", "0xcbd8084f8c759be749340bd20aaed48ec64860e6", "ProxeusFSContract address")
	flag.Bool("devMode", false, "Developers mode")
	flag.String("forceSpp", "http://localhost:8085", "Force spp URL.")
	flag.Bool("autotls", false, "Automatically generate Let's Encrypt certificate (Server must be reachable on port 443 from public internet).")
	flag.Bool("pprof", false, "Pprof debug server")
	flag.String("cfg", ".", "JSON Config file")
	flag.String("blockchainNet", "ropsten", "Blockchain Network (ropsten or mainnet)")
}

func (me Configuration) IsTestMode() bool {
	return strings.ToLower(me.TestMode) == "true"
}

func Setup() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	err := viper.BindEnv("TESTMODE")
	if err != nil {
		log.Println("error bind viper key to a 'testMode' ENV variable")
		return
	}

	err = viper.BindEnv("ETHCLIENTURL")
	if err != nil {
		log.Println("error bind viper key to a 'testMode' ENV variable")
		return
	}

	err = viper.BindEnv("ETHWEBSOCKETURL")
	if err != nil {
		log.Println("error bind viper key to a 'testMode' ENV variable")
		return
	}

	viper.BindPFlags(pflag.CommandLine)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(viper.GetString("cfg")) // optionally look for config in the working directory
	viper.ReadInConfig()                        // Find and read the config file
	viper.Unmarshal(&Config)

	c, _ := json.MarshalIndent(&Config, "", "  ")
	log.Println(string(c))
}

//ETH ChainId for replay protection, see https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md#rationale
func GetChainId() *big.Int {
	if Config.BlockchainNet == "mainnet" {
		return big.NewInt(1)
	}
	return big.NewInt(3) //ropsten
}
