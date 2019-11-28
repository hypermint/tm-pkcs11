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
	pubkeyCmd.Flags().String(FlagKeyLabel, "default", "key label")
	pubkeyCmd.Flags().Bool("show-address", false, "show address")
	rootCmd.AddCommand(pubkeyCmd)
}

var pubkeyCmd = &cobra.Command{
	Use:   "pubkey",
	Short: "Get public key from HSM",
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

		if signer, err := c11ctx.FindKeyPair(nil, []byte(keyLabel)); err != nil {
			return err
		} else if signer == nil {
			return fmt.Errorf("signer is nil")
		} else {
			if pubKey0, ok := signer.Public().(*ecdsa.PublicKey); !ok {
				return fmt.Errorf("not a ECDSA key")
			} else {
				pubKey := signerpv.PublicKeyToPubKeySecp256k1(pubKey0)
				if viper.GetBool("show-address") {
					fmt.Println(pubKey.Address())
				} else {
					fmt.Println(string(cdc.MustMarshalJSON(pubKey)))
				}
				return nil
			}
		}
	},
}
