// Package i18n provides simple map-based internationalization for the CLI.
// Supports "en" (English) and "vi" (Vietnamese) with fallback to English.
package i18n

import "fmt"

var currentLang = "en"

// Init sets the active language. Falls back to "en" if unsupported.
func Init(lang string) {
	if lang == "vi" || lang == "en" {
		currentLang = lang
	} else {
		currentLang = "en"
	}
}

// CurrentLang returns the active language code.
func CurrentLang() string {
	return currentLang
}

// T translates a key to the active language. Supports fmt.Sprintf args.
// Falls back to English if key not found in active language.
// Returns the key itself if not found in any language.
func T(key string, args ...any) string {
	// Try active language first
	if msg, ok := translations[currentLang][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}
	// Fallback to English
	if msg, ok := translations["en"][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}
	// Key not found — return key itself
	return key
}

// translations holds all language strings: lang -> key -> message
var translations = map[string]map[string]string{
	"en": en,
	"vi": vi,
}
