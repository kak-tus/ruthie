package config

import (
	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
	"github.com/kak-tus/ruthie/reader"
	"github.com/kak-tus/ruthie/writer"
)

type Config struct {
	Healthcheck healthcheckConfig
	Reader      reader.Config
	Writer      writer.Config
}

type healthcheckConfig struct {
	Listen string
}

func NewConfig() (*Config, error) {
	fileLdr := fileconf.NewLoader("etc", "/etc")
	envLdr := envconf.NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"file": fileLdr,
				"env":  envLdr,
			},
		},
	)

	configRaw, err := configProc.Load(
		"file:ruthie.yml",
		"env:^RUTHIE_",
	)

	if err != nil {
		return nil, err
	}

	var cnf Config
	if err := conf.Decode(configRaw["ruthie"], &cnf); err != nil {
		return nil, err
	}

	return &cnf, nil
}
