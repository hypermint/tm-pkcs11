package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/hypermint/tm-pkcs11/helpers"
	"github.com/hypermint/tm-pkcs11/signerpv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	genkeyCmd.Flags().String(FlagKeyLabel, "default", "key label")
	rootCmd.AddCommand(genkeyCmd)
}

var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate key on HSM",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		tokenLabel := viper.GetString(FlagTokenLabel)
		password := viper.GetString(FlagPassword)
		hsmSolib := viper.GetString(FlagHsmSolib)
		keyLabel := viper.GetString(FlagKeyLabel)

		c11ctx, err := helpers.CreateCrypto11(hsmSolib, tokenLabel, password, DefaultMaxSessions)
		if err != nil {
			panic(err)
		}

		if err := helpers.GenerateKeyPair2(c11ctx, []byte(keyLabel)); err != nil {
			if err != helpers.ErrKeyFound {
				panic(err)
			}
		}

		if signer, err := c11ctx.FindKeyPair(nil, []byte(keyLabel)); err != nil {
			return err
		} else if signer == nil {
			return fmt.Errorf("signer is nil")
		} else {
			if pubKey0, ok := signer.Public().(*ecdsa.PublicKey); !ok {
				return fmt.Errorf("not a ECDSA key")
			} else {
				pubKey := signerpv.PublicKeyToPubKeySecp256k1(pubKey0)
				fmt.Println(string(cdc.MustMarshalJSON(pubKey)))
				return nil
			}
		}
	},
}
