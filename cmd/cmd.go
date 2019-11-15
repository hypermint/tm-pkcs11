package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func Execute() {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME/.tm-pkcs11")
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	})

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v", err.Error())
		os.Exit(1)
	}
}

