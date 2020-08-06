package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	maddr "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	"github.com/kardianos/service"
)

var daemonLogger service.Logger

type Config struct {
	RendezvousString string
	BootstrapPeers   []maddr.Multiaddr
	ListenAddresses  []maddr.Multiaddr
	ProtocolID       string
}

func main() {
	if err := setupConfig(); err != nil {
		fatalln(err)
	}

	l, err := loadLedger()
	if err != nil {
		fatalln(err)
	}
	ledger = *l
	logln(ledger)

	log.Println("DOCKER_MACHINE_IP: " + os.Getenv("DOCKER_MACHINE_IP"))
	log.Println("PORT: " + os.Getenv("PORT"))
	log.Println("AUTH_SERVER: " + os.Getenv("AUTH_SERVER"))
	if os.Getenv("AUTH_SERVER") == "" {
		log.Fatal("Enviroment AUTH_SERVER does not exist")
	}

	serviceConfig := &service.Config{
		Name:        "golinksDaemon",
		DisplayName: "GoLinks Daemon",
		Description: "golinks daemon",
	}

	d := &daemon{}
	s, err := service.New(d, serviceConfig)
	if err != nil {
		fatalln(err)
	}

	daemonLogger, err = s.Logger(nil)
	if err != nil {
		fatalln(err)
	}

	err = s.Run()
	if err != nil {
		daemonLogger.Error(err)
	}
}

func setupConfig() error {
	daemonHome := HomeDir()

	os.Mkdir(daemonHome, os.ModePerm)

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetDefault("peerPort", 7777)

	viper.AddConfigPath(daemonHome)

	logln("reading config")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		configFile := filepath.Join(daemonHome, "config.json")
		if _, err := os.Create(configFile); err != nil {
			return err
		}
		logln("creating new config file")
		if err := viper.WriteConfig(); err != nil {
			logln("failed to write new config file")
			return err
		}
	}
	return nil
}

var ErrInvalidLedger = errors.New("failed to load ledger from config")

func loadLedger() (*Ledger, error) {
	ledger := &Ledger{}

	ledgerBytes, err := ioutil.ReadFile(filepath.Join(HomeDir(), "ledger.json"))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(ledgerBytes, ledger); err != nil {
		return nil, err
	}

	return ledger, nil
}

func HomeDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".golinks-daemon")
}
