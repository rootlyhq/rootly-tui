package i18n

import (
	"embed"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var localeFS embed.FS

// Language represents a supported language
type Language string

const (
	// LangEnglish is US English
	LangEnglish Language = "en_US"
	// LangFrench is French
	LangFrench Language = "fr_FR"
)

// DefaultLanguage is the fallback language
const DefaultLanguage = LangEnglish

// SupportedLanguages returns all available languages
var SupportedLanguages = []Language{LangEnglish, LangFrench}

// LanguageNames maps language codes to display names
var LanguageNames = map[Language]string{
	LangEnglish: "English (US)",
	LangFrench:  "Francais (FR)",
}

var bundle *i18n.Bundle
var localizer *i18n.Localizer
var currentLang = DefaultLanguage

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load embedded locale files
	_, _ = bundle.LoadMessageFileFS(localeFS, "locales/en_US.yaml")
	_, _ = bundle.LoadMessageFileFS(localeFS, "locales/fr_FR.yaml")

	// Default to English
	localizer = i18n.NewLocalizer(bundle, string(LangEnglish))
}

// SetLanguage sets the current language
func SetLanguage(lang Language) {
	currentLang = lang
	localizer = i18n.NewLocalizer(bundle, string(lang))
}

// GetLanguage returns the current language
func GetLanguage() Language {
	return currentLang
}

// T returns the translation for the given message ID
func T(msgID string) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: msgID,
	})
	if err != nil {
		return msgID
	}
	return msg
}

// Tf returns a formatted translation with template data
func Tf(msgID string, data map[string]interface{}) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: data,
	})
	if err != nil {
		return msgID
	}
	return msg
}

// ListLanguages returns language codes for selector
func ListLanguages() []string {
	result := make([]string, len(SupportedLanguages))
	for i, lang := range SupportedLanguages {
		result[i] = string(lang)
	}
	return result
}

// LanguageName returns the display name for a language code
func LanguageName(code string) string {
	if name, ok := LanguageNames[Language(code)]; ok {
		return name
	}
	return code
}

// LanguageIndex returns the index of the given language code
func LanguageIndex(code string) int {
	for i, lang := range SupportedLanguages {
		if string(lang) == code {
			return i
		}
	}
	return 0 // Default to first language
}

// DetectLanguage detects the system locale and returns a supported language.
// Falls back to DefaultLanguage (en_US) if not supported.
func DetectLanguage() Language {
	// Check LANG and LC_ALL environment variables
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(envVar); locale != "" {
			// Locale format: en_US.UTF-8 or fr_FR.UTF-8
			// Strip encoding suffix
			locale = strings.Split(locale, ".")[0]

			// Check if exact match
			for _, lang := range SupportedLanguages {
				if string(lang) == locale {
					return lang
				}
			}

			// Check language prefix (e.g., "fr" matches "fr_FR")
			langPrefix := strings.Split(locale, "_")[0]
			for _, lang := range SupportedLanguages {
				if strings.HasPrefix(string(lang), langPrefix+"_") {
					return lang
				}
			}
		}
	}

	return DefaultLanguage
}
