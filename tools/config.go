package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/ncruces/zenity"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const APP_SERVICE = "KasaSysTray"
const KEYRING_KEY = "passphrase"

type Configuration struct {
	Passphrase    string `json:"passphrase"`
	EncryptedAuth string `json:"encrypted_auth"`
	AutoConnect   bool   `json:"auto_connect"`
}

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func SetupConfiguration() *Configuration {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")

	viper.AutomaticEnv()

	var configuration Configuration

	if err := viper.ReadInConfig(); err != nil {
		viper.SetConfigFile(path.Join("./", "config.json"))
		log.Printf("Error reading config file, %s\n", err)
	}

	settings := viper.AllSettings()

	transcode(settings, &configuration)

	return &configuration
}

func (config *Configuration) WriteConfiguration() error {
	viper.Set("encrypted_auth", config.EncryptedAuth)
	viper.Set("auto_connect", config.AutoConnect)

	log.Println("\nWriting configuration...", viper.ConfigFileUsed())

	err := viper.WriteConfig()
	if err != nil {
		log.Printf("Error writing config file, %s\n", err)
		return err
	}
	return nil
}

func (config *Configuration) DeleteConfig() error {
	fpath := viper.ConfigFileUsed()
	return os.Remove(fpath)
}

func (config *Configuration) SetPassphrase(passphrase string) error {
	config.Passphrase = passphrase
	return nil
}

func (config *Configuration) SetAuth(username string, password string) error {
	auth := Auth{Username: username, Password: password}
	err := config.encrypt(auth)
	if err != nil {
		return err
	}
	return config.WriteConfiguration()
}

func (config *Configuration) ReadAuth(useGUI bool) (*Auth, bool, error) {
	var auth Auth
	var isFresh bool = false
	if config.EncryptedAuth == "" {
		if useGUI {
			username, password, err := SetAuthGUI()
			if err != nil {
				return nil, isFresh, err
			}
			auth = Auth{Username: username, Password: password}
		} else {
			// Let's get the auth from user
			fmt.Println("\nUsername:")
			fmt.Scanln(&auth.Username)
			fmt.Println("\nPassword:")
			fmt.Scanln(&auth.Password)
		}

		err := config.encrypt(auth)
		if err != nil {
			return nil, isFresh, err
		}

		isFresh = true
	}
	ciphertext, err := base64.StdEncoding.DecodeString(config.EncryptedAuth)
	if err != nil {
		return nil, isFresh, err
	}
	data, err := config.decrypt(ciphertext)
	if err != nil {
		return nil, isFresh, err
	}
	err = json.Unmarshal(data, &auth)
	if err != nil {
		return nil, isFresh, err
	}
	return &auth, isFresh, nil
}

func (config *Configuration) encrypt(data interface{}) error {
	// Use DES to encrypt the data.
	dataByte, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c, err := aes.NewCipher([]byte(createHash(config.Passphrase)))
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nonce, nonce, dataByte, nil)

	config.EncryptedAuth = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

func (config *Configuration) decrypt(data []byte) ([]byte, error) {
	c, err := aes.NewCipher([]byte(createHash(config.Passphrase)))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	data, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(in)
	json.NewDecoder(buf).Decode(out)
}

func SetPassphraseGUI() (passphrase string, err error) {
	passphrase, err = keyring.Get(APP_SERVICE, KEYRING_KEY)

	if err != nil {
		_, passphrase, err = zenity.Password(
			zenity.Title("Set your Keyring Passphrase"),
		)
		if err != nil {
			return "", err
		}
		err = keyring.Set(APP_SERVICE, KEYRING_KEY, passphrase)
		if err != nil {
			return "", err
		}
	}

	return passphrase, nil
}

func SetAuthGUI() (username string, password string, err error) {
	username, err = zenity.Entry(
		"Enter your Email ID",
		zenity.Title("Set your Kasa Credentials"),
	)
	if err != nil {
		return "", "", err
	}
	_, password, err = zenity.Password(
		zenity.Title("Set your Kasa Credentials"),
	)
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}

func ResetKeyring() error {
	err := keyring.Delete(APP_SERVICE, KEYRING_KEY)
	if err != nil {
		return err
	}
	return nil
}

func ResetAll(config *Configuration) error {
	err := config.DeleteConfig()
	if err != nil {
		return err
	}
	err = ResetKeyring()
	if err != nil {
		return err
	}
	return nil
}

func DisplayErrorGUI(err error) {
	zenity.Error(
		err.Error(),
		zenity.Title("Error"),
	)
}

func Notify(title string, message string, icon zenity.DialogIcon) error {
	return zenity.Notify(
		message,
		zenity.Title(title),
		icon,
	)
}
