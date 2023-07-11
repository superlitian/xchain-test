package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

// Config the all config
type Config struct {
	ServiceCfg  *ServiceConfig
	AccountCfg  *AccountConfig
	XchainCfg   *XchainConfig
	ContractCfg *ContractConfig
}

// ServiceConfig the service config
type ServiceConfig struct {
	Times    int `yaml:"times,omitempty"`
	Routines int `yaml:"routines,omitempty"`
}

// AccountConfig the account config
type AccountConfig struct {
	Nums int `yaml:"nums,omitempty"`
}

type XchainConfig struct {
	IsSingle                    bool     `yaml:"isSingle,omitempty"`
	Nodes                       []string `yaml:"nodes,omitempty"`
	IsNeedTransferToAdmin       bool     `yaml:"isNeedTransferToAdmin,omitempty"`
	IsNeedCreateContractAccount bool     `yaml:"isNeedCreateContractAccount,omitempty"`
	ContractAccount             string   `yaml:"contractAccount,omitempty"`
}

type ContractConfig struct {
	Module     string            `yaml:"module,omitempty"`
	Name       string            `yaml:"name,omitempty"`
	Initialize map[string]string `yaml:"initialize,omitempty"`
	Method     string            `yaml:"method,omitempty"`
	Params     map[string]string `yaml:"params,omitempty"`
}

const confPath = "./conf"

var ComCfg Config

func LoadConfig() (*Config, error) {
	// service conf
	name := "service.yaml"
	realPath := filepath.Join(confPath, name)
	serviceCfg := &ServiceConfig{}
	yamlFile, err := ioutil.ReadFile(realPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, serviceCfg)
	if err != nil {
		return nil, err
	}
	// account conf
	name = "account.yaml"
	realPath = filepath.Join(confPath, name)
	accountCfg := &AccountConfig{}
	yamlFile, err = ioutil.ReadFile(realPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, accountCfg)
	if err != nil {
		return nil, err
	}

	// xchain conf
	name = "xchain.yaml"
	realPath = filepath.Join(confPath, name)
	xchainCfg := &XchainConfig{}
	yamlFile, err = ioutil.ReadFile(realPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, xchainCfg)
	if err != nil {
		return nil, err
	}

	// contract conf
	name = "contract.yaml"
	realPath = filepath.Join(confPath, name)
	contractCfg := &ContractConfig{}
	yamlFile, err = ioutil.ReadFile(realPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, contractCfg)
	if err != nil {
		return nil, err
	}

	ComCfg.ServiceCfg = serviceCfg
	ComCfg.AccountCfg = accountCfg
	ComCfg.XchainCfg = xchainCfg
	ComCfg.ContractCfg = contractCfg
	return &Config{
		ServiceCfg:  serviceCfg,
		AccountCfg:  accountCfg,
		XchainCfg:   xchainCfg,
		ContractCfg: contractCfg,
	}, nil
}
