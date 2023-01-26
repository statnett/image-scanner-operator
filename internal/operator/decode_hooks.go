package operator

import (
	"reflect"
	"regexp"

	"github.com/mitchellh/mapstructure"
)

// stringToRegexpHookFunc returns a DecodeHookFunc that converts strings to regexp.Regexp.
func stringToRegexpHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(&regexp.Regexp{}) {
			return data, nil
		}

		return regexp.Compile(data.(string))
	}
}
