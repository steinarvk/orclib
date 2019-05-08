package orcmain

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/steinarvk/orclib/lib/versioninfo"
)

func Init(programName string, rootCommand *cobra.Command) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	var configFilename string

	rootCommand.PersistentFlags().StringVar(&configFilename, "config", "", "config filename")

	cobra.OnInitialize(func() {
		if configFilename != "" {
			viper.SetConfigFile(configFilename)
		} else {
			home, err := homedir.Dir()
			if err != nil {
				logrus.Fatalf("Failed to find home directory: %v", err)
			}

			viper.AddConfigPath(filepath.Join(home, fmt.Sprintf(".config/%s", programName)))
			viper.SetConfigName(fmt.Sprintf("%s.yaml", programName))
		}

		viper.AutomaticEnv()
		if err := viper.ReadInConfig(); err != nil {
			_, ok := err.(viper.ConfigFileNotFoundError)
			if !ok {
				logrus.Fatalf("Failed to read config file %q: %v", configFilename, err)
			}
		} else {
			logrus.Infof("Using config file %q", viper.ConfigFileUsed())
		}

		logrus.WithFields(versioninfo.MakeFields()).Infof("Starting %s", programName)
	})
}

func Main(rootCommand *cobra.Command) {
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
