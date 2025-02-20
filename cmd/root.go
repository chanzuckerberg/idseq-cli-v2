package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/chanzuckerberg/czid-cli/cmd/amr"
	"github.com/chanzuckerberg/czid-cli/cmd/consensusGenome"
	"github.com/chanzuckerberg/czid-cli/cmd/generateMetadataTemplate"
	"github.com/chanzuckerberg/czid-cli/cmd/metagenomics"
	"github.com/chanzuckerberg/czid-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "czid",
	Short: "A CLI for uploading samples to Chan Zuckerberg ID",
}

func init() {
	RootCmd.AddCommand(metagenomics.MetagenomicsCmd)
	RootCmd.AddCommand(consensusGenome.ConsensusGenomeCmd)
	RootCmd.AddCommand(amr.AmrCmd)
	RootCmd.AddCommand(generateMetadataTemplate.GenerateMetadataTemplateCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Print verbose logs")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := util.GetConfigDir()
		if err != nil {
			log.Fatal(err)
		}
		viper.SetConfigFile(path.Join(configDir, "config.yaml"))
	}

	// Check if verbose flag is set
	verbose, _ := RootCmd.Flags().GetBool("verbose")

	viper.SetEnvPrefix("czid_cli")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Error reading config file: %s\n", err.Error())
		} else if verbose {
			fmt.Println("Config file not found. Run the command with '--config' to explicitly pass in a config.yaml file. Or set environment variables with the prefix CZID_CLI_* (eg CZID_CLI_SECRET)")
		}

	}
}
