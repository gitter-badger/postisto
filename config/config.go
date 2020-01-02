package config

import (
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Accounts map[string]struct {
		Enabled bool `yaml:"enabled"`
		Server string `yaml:"server"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		SortMailbox string `yaml:"sort_mailbox"`
		IMAPS bool `yaml:"imaps"`
		Starttls bool `yaml:"starttls"`
	} `yaml:"accounts"`
	Filters map[string]map[string]struct {
		Commands []map[string]interface{} `yaml:"commands"`
		Rules    []map[string]interface{} `yaml:"rules"`
	} `yaml:"filters"`
	Settings struct {
		//logging
		UseGPGAgent bool `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

func GetConfig(configPath string) (Config, error) {

	cfg := Config{}
	var configFiles []string

	stat, err := os.Stat(configPath)
	if err != nil {
		return cfg, err
	}

	if stat.IsDir() {
		err := filepath.Walk(configPath,
			func(path string, info os.FileInfo, err error) error {
				//log.Println(path)
				if err != nil {
					return err
				}
				if stat, _ := os.Stat(path); !stat.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
					configFiles = append(configFiles, path)
				}
				return nil
			})
		if err != nil {
			return cfg, err
		}
	} else {
		configFiles = append(configFiles, configPath)
	}

	for _, file := range configFiles {
		//log.Printf(">>> %v", file)
		fileCfg := Config{}
		yamlFile, err := ioutil.ReadFile(file)

		if err != nil {
			//log.Fatalf("Error reading config file %v: %v", file, err)
			return cfg, err
		}

		err = yaml.Unmarshal(yamlFile, &fileCfg)

		if err != nil {
			//log.Fatalf("Error unmarshing config file %v: %v", file, err)
			return cfg, err
		}

		if err := mergo.Merge(&cfg, fileCfg, mergo.WithOverride); err != nil {
			log.Fatalf("yooo %v", err)
			return cfg, err
		}
	}

	//TODO validate our user's config

	return cfg, nil
}
