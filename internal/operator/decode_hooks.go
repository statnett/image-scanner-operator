package operator

import (
	"reflect"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
)

// stringToRegexpHookFunc returns a DecodeHookFunc that converts strings to regexp.Regexp.
func stringToRegexpHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeFor[*regexp.Regexp]() {
			return data, nil
		}

		return regexp.Compile(data.(string))
	}
}
