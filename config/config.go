package config

import (
	"bytes"
	_ "embed"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path"
	"runtime"
)

//go:embed config.yaml
var configBytes []byte

var _config Config

type LogConfig struct {
	Dir        string `mapstructure:"dir"`
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type MysqlConfig struct {
	Url             string `mapstructure:"url"`
	Prefix          string `mapstructure:"prefix"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	SlowThreshold   int    `mapstructure:"slow_threshold"`
}

type IndexConfig struct {
	ScanBlock uint64 `mapstructure:"scan_block"`
}

type AppConfig struct {
	Name           string `mapstructure:"name"`
	Port           int    `mapstructure:"port"`
	RoutePrefix    string `mapstructure:"route_prefix"`
	MintStartBlock uint64 `mapstructure:"mint_start_block"`
	MintEndBlock   uint64 `mapstructure:"mint_end_block"`
	BridgeEvmBlock uint64 `mapstructure:"bridge_evm_block"`
}

type Config struct {
	App      AppConfig   `yaml:"fans_index"`
	BscIndex IndexConfig `yaml:"bsc"`
	Log      LogConfig   `yaml:"log"`
	Mysql    MysqlConfig `yaml:"mysql"`
}

func GetConfig() Config {
	return _config
}

func init() {
	var _conf Config

	conf := viper.New()
	conf.SetConfigType("yaml")

	if err := conf.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
		panic(err)
	}

	_, filename, _, _ := runtime.Caller(0)

	bscBytes, err := ioutil.ReadFile(path.Join(path.Dir(filename), "bsc.yaml"))
	if err != nil {
		panic(err)
	}

	{
		bscConf := viper.New()
		bscConf.SetConfigType("yaml")
		if err := bscConf.ReadConfig(bytes.NewBuffer(bscBytes)); err != nil {
			panic(err)
		}
		if err := bscConf.Sub("bsc").Unmarshal(&_conf.BscIndex); err != nil {
			panic(err)
		}
	}

	{
		logConf := conf.Sub("log")
		if err := logConf.Unmarshal(&_conf.Log); err != nil {
			panic(err)
		}
	}

	{
		mysqlConf := conf.Sub("mysql")
		if err := mysqlConf.Unmarshal(&_conf.Mysql); err != nil {
			panic(err)
		}
	}

	{
		appConf := conf.Sub("fans_index")
		if err := appConf.Unmarshal(&_conf.App); err != nil {
			panic(err)
		}
	}

	_config = _conf
}

func SaveBSCConfig(blockNumber uint64) error {
	bsc := _config.BscIndex
	bsc.ScanBlock = blockNumber + 1

	type config struct {
		Bsc IndexConfig `yaml:"bsc"`
	}

	_config := config{Bsc: bsc}
	data, err := yaml.Marshal(_config)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile("./config/bsc.yaml", data, 0666); err != nil {
		return err
	}

	return nil
}
