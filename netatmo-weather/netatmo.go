package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/exzz/netatmo-api-go"
)

const (
	indoorModuleType  = "NAModule4"
	outdoorModuleType = "NAModule1"
	mainModuleType    = "NAMain"
)

// API credentials
type NetatmoConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	Username     string `toml:"username"`
	Password     string `toml:"password"`
}

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Printf("Cannot parse config file: %s\n", err)
		os.Exit(1)
	}

	n, err := netatmo.NewClient(netatmo.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Username:     config.Username,
		Password:     config.Password,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dc, err := n.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, station := range dc.Stations() {
		var mainModule *netatmo.Device
		var outdoorModules []*netatmo.Device
		var indoorModules []*netatmo.Device

		for _, module := range station.Modules() {
			switch module.Type {
			case mainModuleType:
				mainModule = module
			case outdoorModuleType:
				outdoorModules = append(outdoorModules, module)
			case indoorModuleType:
				indoorModules = append(indoorModules, module)
			}
		}
		var head = ":deciduous_tree: "

		if len(outdoorModules) > 0 {
			module := outdoorModules[0]
			temperature, temperatureAdded := getTemperature(module)
			if !temperatureAdded {
				head += "N/A"
			} else {
				head += temperature
			}
		} else {
			head += "N/A"
		}
		head += " \\ "

		if mainModule != nil {
			head += ":house: "
			temperature, temperatureAdded := getTemperature(mainModule)
			if !temperatureAdded {
				head += "N/A"
			} else {
				head += temperature
			}
		}
		fmt.Println(head)
		fmt.Println("---")

		ct := time.Now().UTC().Unix()

		allModules := append(append([]*netatmo.Device{mainModule}, outdoorModules...), indoorModules...)

		for _, module := range allModules {

			fmt.Printf("Module : %s (%s)\n", module.ModuleName, getDisplayType(module))

			if module.DashboardData.LastMeasure == nil {
				fmt.Printf("Skipping %s, no measurement data available.\n", module.ModuleName)
				continue
			}
			ts, data := module.Info()
			for dataName, value := range data {
				fmt.Printf("%s : %v (updated %ds ago)\n", dataName, value, ct-ts)
			}

			ts, data = module.Data()
			for dataName, value := range data {
				fmt.Printf("%s : %v (updated %ds ago)\n", dataName, value, ct-ts)
			}
			fmt.Println("---")
		}
	}
}

func getConfigFilePath() string {
	info, err := os.Stat("config.toml")
	if err == nil && !info.IsDir() {
		return "config.toml"
	}
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	path := filepath.Join(usr.HomeDir, ".config", "netatmo", "config.toml")
	info, err = os.Stat(path)
	if err == nil && !info.IsDir() {
		return path
	}
	return ""
}

func readConfig() (*NetatmoConfig, error) {
	configFilePath := getConfigFilePath()
	if configFilePath == "" {
		return nil, errors.New("no config file found")
	}
	var config NetatmoConfig
	if _, err := toml.DecodeFile(configFilePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getDisplayType(module *netatmo.Device) string {
	switch module.Type {
	case mainModuleType:
		return ":grey_exclamation: Main"
	case outdoorModuleType:
		return ":deciduous_tree: Outdoor"
	case indoorModuleType:
		return ":house: Indoor"
	default:
		return ":grey_question: Unknown"
	}
}

func getTemperature(module *netatmo.Device) (string, bool) {

	_, data := module.Data()
	if temperature, ok := data["Temperature"]; ok {
		if t, castOk := temperature.(string); castOk {
			return t + " °C", true
		}
		if t, castOk := temperature.(float64); castOk {
			return fmt.Sprintf("%.1f °C", t), true
		}
		if t, castOk := temperature.(float32); castOk {
			return fmt.Sprintf("%.1f °C", t), true
		}
		if t, castOk := temperature.(int32); castOk {
			return fmt.Sprintf("%d °C", t), true
		}
		if t, castOk := temperature.(int64); castOk {
			return fmt.Sprintf("%d °C", t), true
		}
	}

	return "", false
}
