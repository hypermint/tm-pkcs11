package cmd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/privval"
	"os"
)

const flagKeyFile = "key-file"
const flagStateFile = "state-file"
const flagOutputFile = "output-file"
const flagOutputFormat = "output-format"
const flagPemType = "pem-type"

func init() {
	rootCmd.AddCommand(pkcs8Cmd)
	pkcs8Cmd.Flags().String(flagKeyFile, "", "key file")
	pkcs8Cmd.Flags().String(flagStateFile, "", "state file")
	pkcs8Cmd.Flags().String(flagOutputFormat, "pem", "output format")
	pkcs8Cmd.Flags().String(flagOutputFile, "", "output file path")
	pkcs8Cmd.Flags().String(flagPemType, "PRIVATE KEY", "output file path")
}

var pkcs8Cmd = &cobra.Command{
	Use:   "pkcs8",
	Short: "Convert private key to pkcs8",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		keyFile := viper.GetString(flagKeyFile)
		stateFile := viper.GetString(flagStateFile)

		filePV := privval.LoadFilePVEmptyState(keyFile, stateFile)
		privKey := filePV.Key.PrivKey

		var privateKey *ecdsa.PrivateKey
		switch privKey := privKey.(type) {
		case secp256k1.PrivKeySecp256k1:
			key, _  := btcec.PrivKeyFromBytes(btcec.S256(), privKey[:])
			privateKey = (*ecdsa.PrivateKey)(key)
		default:
			return errors.New("invalid private key")
		}

		pkcs8, err := MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return err
		}
		switch viper.GetString(flagOutputFormat) {
		case "pem":
			pemType := viper.GetString(flagPemType)
			pkcs8 = pem.EncodeToMemory(&pem.Block{
				Type: pemType,
				Bytes: pkcs8,
			})
		case "der":
		default:
			return errors.New("unknown output format")
		}

		outputPath := viper.GetString(flagOutputFile)
		if len(outputPath) > 0 {
			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = file.Write(pkcs8)
			if err != nil {
				return err
			}
		} else {
			os.Stdout.Write(pkcs8)
		}
		return nil
	},
}

// see crypto/x509/pkcs8.go

var oidPublicKeyECDSA   = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}

type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}

func MarshalPKCS8PrivateKey(key interface{}) ([]byte, error) {
	var privKey pkcs8

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		oid, ok := oidFromNamedCurve(k.Curve)
		if !ok {
			return nil, errors.New("x509: unknown curve while marshaling to PKCS#8")
		}

		oidBytes, err := asn1.Marshal(oid)
		if err != nil {
			return nil, errors.New("x509: failed to marshal curve OID: " + err.Error())
		}

		privKey.Algo = pkix.AlgorithmIdentifier{
			Algorithm: oidPublicKeyECDSA,
			Parameters: asn1.RawValue{
				FullBytes: oidBytes,
			},
		}

		if privKey.PrivateKey, err = marshalECPrivateKeyWithOID(k, nil); err != nil {
			return nil, errors.New("x509: failed to marshal EC private key while building PKCS#8: " + err.Error())
		}
	default:
		return nil, fmt.Errorf("x509: unknown key type while marshaling PKCS#8: %T", key)
	}

	return asn1.Marshal(privKey)
}

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

func marshalECPrivateKeyWithOID(key *ecdsa.PrivateKey, oid asn1.ObjectIdentifier) ([]byte, error) {
	privateKeyBytes := key.D.Bytes()
	paddedPrivateKey := make([]byte, (key.Curve.Params().N.BitLen()+7)/8)
	copy(paddedPrivateKey[len(paddedPrivateKey)-len(privateKeyBytes):], privateKeyBytes)

	return asn1.Marshal(ecPrivateKey{
		Version:       1,
		PrivateKey:    paddedPrivateKey,
		NamedCurveOID: oid,
		PublicKey:     asn1.BitString{Bytes: elliptic.Marshal(key.Curve, key.X, key.Y)},
	})
}

var (
	oidNamedCurveP224 = asn1.ObjectIdentifier{1, 3, 132, 0, 33}
	oidNamedCurveP256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	oidNamedCurveP384 = asn1.ObjectIdentifier{1, 3, 132, 0, 34}
	oidNamedCurveP521 = asn1.ObjectIdentifier{1, 3, 132, 0, 35}
	oidNamedCurveSecp256k1 = asn1.ObjectIdentifier{1, 3, 132, 0, 10}
)

func oidFromNamedCurve(curve elliptic.Curve) (asn1.ObjectIdentifier, bool) {
	switch curve {
	case elliptic.P224():
		return oidNamedCurveP224, true
	case elliptic.P256():
		return oidNamedCurveP256, true
	case elliptic.P384():
		return oidNamedCurveP384, true
	case elliptic.P521():
		return oidNamedCurveP521, true
	case btcec.S256():
		return oidNamedCurveSecp256k1, true
	}

	return nil, false
}
