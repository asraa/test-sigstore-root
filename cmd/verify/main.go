package main

import (
	"crypto/elliptic"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asraa/sigstore-root/cmd/tuf/app"
	"github.com/asraa/sigstore-root/pkg/keys"
	"github.com/theupdateframework/go-tuf"
	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/verify"
)

type file string

func (f *file) String() string {
	return string(*f)
}

func (f *file) Set(s string) error {
	if s == "" {
		return errors.New("flag must be specified")
	}
	if _, err := os.Stat(filepath.Clean(s)); os.IsNotExist(err) {
		return err
	}
	*f = file(s)
	return nil
}

func toCert(filename string) (*x509.Certificate, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return keys.ToCert(fileBytes)
}

// Map from Key ID to Signing Key
type KeyMap map[string]*keys.SigningKey

func getKeyID(key keys.SigningKey) (*string, error) {
	pub := key.PublicKey
	pk := &data.Key{
		Type:       data.KeyTypeECDSA_SHA2_P256,
		Scheme:     data.KeySchemeECDSA_SHA2_P256,
		Algorithms: data.KeyAlgorithms,
		Value:      data.KeyValue{Public: elliptic.Marshal(pub.Curve, pub.X, pub.Y)},
	}
	if len(pk.IDs()) == 0 {
		return nil, errors.New("error getting key ID")
	}
	return &pk.IDs()[0], nil
}

func signingKeyFromDir(dirname string) (*keys.SigningKey, error) {
	// Expect *_device_cert.pem, *_key_cert.pem, *_pubkey.pem in each key directory.
	serialStr := filepath.Base(dirname)
	serial, err := strconv.Atoi(serialStr)
	if err != nil {
		return nil, fmt.Errorf("invalid key directory name %s: %s", dirname, err)
	}

	var pubKey []byte
	var deviceCert []byte
	var keyCert []byte
	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("panic accessing path %q: %v\n", path, err)
			return err
		}
		if strings.HasSuffix(info.Name(), "_pubkey.pem") {
			pubKey, err = ioutil.ReadFile(path)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(info.Name(), "_key_cert.pem") {
			keyCert, err = ioutil.ReadFile(path)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(info.Name(), "_device_cert.pem") {
			deviceCert, err = ioutil.ReadFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys.ToSigningKey(serial, pubKey, deviceCert, keyCert)
}

func verifySigningKeys(dirname string, rootCA *x509.Certificate) (*KeyMap, error) {
	// Get all signing keys in the directory.
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	keyMap := make(KeyMap)
	for _, file := range files {
		if file.IsDir() {
			key, err := signingKeyFromDir(filepath.Join(dirname, file.Name()))
			if err != nil {
				return nil, err
			}
			if err = key.Verify(rootCA); err != nil {
				log.Printf("error verifying key %d: %s", key.SerialNumber, err)
				return nil, err
			} else {
				log.Printf("verified key %d", key.SerialNumber)
			}

			id, err := getKeyID(*key)
			if err != nil {
				return nil, err
			}
			keyMap[*id] = key
		}
	}
	return &keyMap, nil
}

func verifyMetadata(repository string, keys KeyMap) error {
	// logs the state of each metadata file, including number of signatures to achieve threshold
	// and verifies the signatures in each file.
	store := tuf.FileSystemStore(repository, nil)
	db, err := app.CreateDb(store)
	if err != nil {
		return err
	}
	root, err := app.GetRootFromStore(store)
	if err != nil {
		return err
	}

	for name, role := range root.Roles {
		log.Printf("\nVerifying %s...", name)
		signed, err := app.GetSignedMeta(store, name+".json")
		if err != nil {
			return err
		}
		if err = db.VerifySignatures(signed, name); err != nil {
			if _, ok := err.(verify.ErrRoleThreshold); ok {
				// we may not have all the sig, allow partial sigs
				log.Printf("\tContains %d/%d valid signatures", err.(verify.ErrRoleThreshold).Actual, role.Threshold)
			} else {
				log.Printf("\tError verifying: %s", err)
			}
		} else {
			log.Printf("\tSuccess! Signatures valid and threshold achieved")
		}
	}

	return nil
}

func main() {
	log.SetFlags(0)
	var fileFlag file
	flag.Var(&fileFlag, "root", "Yubico root certificate")
	keyDir := flag.String("key-directory", "../../../ceremony/2021-05-03/keys", "Directory with key products")
	repository := flag.String("repository", "", "path to repository")
	// TODO: Add path to repository to verify metadata.
	flag.Parse()

	rootCA, err := toCert(string(fileFlag))
	if err != nil {
		log.Printf("failed to parse root CA: %s", err)
		os.Exit(1)
	}

	keyMap, err := verifySigningKeys(*keyDir, rootCA)
	if err != nil {
		log.Printf("error verifying signing keys: %s", err)
		os.Exit(1)
	}

	if *repository != "" {
		if err := verifyMetadata(*repository, *keyMap); err != nil {
			log.Printf("error verifying signing keys: %s", err)
			os.Exit(1)
		}
	}
}
