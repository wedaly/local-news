package i18n

import (
	"encoding/xml"
	"io"
	"os"
	"path"
)

// Config is a locale-specific configuration for the program
type Config struct {
	FormLabelColor            string `xml:"colors>formLabel"`
	FormButtonBackgroundColor string `xml:"colors>formButtonBackground"`
	FormButtonTextColor       string `xml:"colors>formButtonText"`
	FormFieldBackgroundColor  string `xml:"colors>formFieldBackground"`
	FormFieldTextColor        string `xml:"colors>formFieldText"`
	FormErrorBackgroundColor  string `xml:"colors>formErrorBackground"`
	FormErrorTextColor        string `xml:"colors>formErrorText"`
	ModalTextColor            string `xml:"colors>modalText"`
	ModalBackgroundColor      string `xml:"colors>modalBackground"`
}

// DefaultConfig returns a default configuration, used in case
// a locale-specific configuration cannot be found.
func DefaultConfig() Config {
	return Config{
		FormLabelColor:            "yellow",
		FormButtonBackgroundColor: "blue",
		FormButtonTextColor:       "white",
		FormFieldBackgroundColor:  "blue",
		FormFieldTextColor:        "white",
		FormErrorBackgroundColor:  "red",
		FormErrorTextColor:        "white",
		ModalBackgroundColor:      "blue",
		ModalTextColor:            "black",
	}
}

// ParseConfigFromXml loads a configuration from XML.
func ParseConfigXml(r io.Reader) Config {
	decoder := xml.NewDecoder(r)
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return DefaultConfig()
	} else {
		return config
	}
}

// LoadConfig locates and loads a locale-specific configuration file.
// Search paths are directories to search in order.
// When an XML file is found at {SEARCHPATH}/{LOCALE}/config.xml,
// it is parsed and returned.
// If no locale-specific config can be found, then the default config is returned.
func LoadConfig(searchPaths []string) Config {
	locale := GetMessageLocaleFromEnv()
	for _, dir := range searchPaths {
		path := path.Join(dir, locale, "config.xml")
		if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
			if f, err := os.Open(path); err == nil {
				return ParseConfigXml(f)
			}
		}
	}
	return DefaultConfig()
}
