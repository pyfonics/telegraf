//go:generate ../../../tools/readme_config_includer/generator
package units

import (
	_ "embed"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	un "github.com/bcicen/go-units"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

//var unitMap map[string]interface{}

type Units struct {
	Pattern     string            `toml:"pattern"`
	From        string            `toml:"from"`
	Replacement string            `toml:"replacement"`
	To          string            `toml:"to"`
	IsCounter   bool              `toml:"is_counter"`
	AutoSuffix  bool              `toml:"unit_suffix"`
	Tags        map[string]string `toml:"tags"`
	Log         telegraf.Logger   `toml:"-"`
	FromUnit    un.Unit           `toml:"-`
	ToUnit      un.Unit           `toml:"-`
	UnitSuffix  string            `toml:"-`
	ReCompile   *regexp.Regexp    `toml:"-"`
}

func (*Units) SampleConfig() string {
	return sampleConfig
}

func (u *Units) Init() error {

	// Check if the unit exists
	unit, err := un.Find(u.From)
	if err != nil {
		return fmt.Errorf("Cannot find the unit: %q", u.From)
	} else {
		u.FromUnit = unit
	}

	// Check if the destination unit exists
	dest, err := un.Find(u.To)
	if err != nil {
		return fmt.Errorf("Cannot find the destination unit: %q", u.To)
	} else {
		u.ToUnit = dest
		u.UnitSuffix = strings.ToLower(u.To)
	}

	// Set default value for replacement string if not passed
	if u.Replacement == "" {
		u.Replacement = "${0}"
	}

	// Check that a regex pattern is passed
	if u.Pattern == "" {
		return errors.New("A valid regex pattern must be passed")
	}

	// Check that the regex compiles
	re, err := regexp.Compile(u.Pattern)
	if err != nil {
		return errors.New("Invalid regex pattern, did not compile")
	} else {
		u.ReCompile = re
	}

	return nil
}

func (u *Units) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for key, value := range metric.Fields() {
			if fieldMatched := u.ReCompile.MatchString(key); fieldMatched {

				fieldName := u.ReCompile.ReplaceAllString(key, u.Replacement)

				val, e := toFloat(value)
				if e == false {
					u.Log.Errorf("Cannot convert field to float unit: %v", key)
				} else {

					convertedValue, err := un.ConvertFloat(val, u.FromUnit, u.ToUnit)

					if err != nil {
						u.Log.Errorf("Cannot convert unit: %v", err)
					} else {

						if u.AutoSuffix {
							if !strings.HasSuffix(fieldName, u.UnitSuffix) {
								fieldName += "_" + u.UnitSuffix
							}
						}

						if u.IsCounter {
							if !strings.HasSuffix(fieldName, "_total") {
								fieldName += "_total"
							}
						}

						metric.RemoveField(key)
						metric.AddField(fieldName, convertedValue.Float())
					}

				}
			}

			for key, value := range u.Tags {
				metric.AddTag(key, value)
			}
		}
	}

	return metrics
}

// isHexadecimal directly copied from processors/converter
func isHexadecimal(value string) bool {
	return len(value) >= 3 && strings.ToLower(value)[1] == 'x'
}

// parseHexadecimal directly copied from processors/converter
func parseHexadecimal(value string) (float64, error) {
	i := new(big.Int)

	_, success := i.SetString(value, 0)
	if !success {
		return 0, errors.New("unable to parse string to big int")
	}

	f := new(big.Float).SetInt(i)
	result, _ := f.Float64()

	return result, nil
}

// toFloat directly copied from processors/converter
func toFloat(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	case float64:
		return value, true
	case bool:
		if value {
			return 1.0, true
		}
		return 0.0, true
	case string:
		if isHexadecimal(value) {
			result, err := parseHexadecimal(value)
			return result, err == nil
		}

		result, err := strconv.ParseFloat(value, 64)
		return result, err == nil
	}
	return 0.0, false
}

func init() {
	processors.Add("units", func() telegraf.Processor {
		return &Units{}
	})
}
