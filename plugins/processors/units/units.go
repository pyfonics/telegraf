//go:generate ../../../tools/readme_config_includer/generator
package units

import (
	_ "embed"
	"errors"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/martinlindhe/unit"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

//var unitMap map[string]interface{}

type Units struct {
	Pattern          string                 `toml:"pattern"`
	Unit             string                 `toml:"unit"`
	Replacement      string                 `toml:"replacement"`
	DestUnit         string                 `toml:"dest_unit"`
	IsCounter        bool                   `toml:"is_counter"`
	AutoSuffix       bool                   `toml:"auto_suffix"`
	Log              telegraf.Logger        `toml:"-"`
	AllUnits         map[string]interface{} `toml:"-"`
	TemperatureUnits []string               `toml:"-"`
	IsTemp           bool                   `toml:"-"`
	IsValidUnit      bool                   `toml:"-"`
	ReCompile        *regexp.Regexp         `toml:"-"`
}

func (*Units) SampleConfig() string {
	return sampleConfig
}

// TODO: change all receivers to u

func (u *Units) Init() error {
	u.AllUnits = allUnits()
	u.TemperatureUnits = temperatureUnits()

	// default value for replacement string if not passed
	if u.Replacement == "" {
		u.Replacement = "${0}"
	}

	// TODO: add check if unit exists validation here
	_, ok := u.AllUnits[u.Unit]

	// check if declared unit is a temperature type
	u.IsTemp = false
	for _, b := range u.TemperatureUnits {
		if b == u.Unit {
			u.IsTemp = true
		}
	}

	// check if the declared unit if exists
	if !(ok || u.IsTemp) {
		return errors.New("The unit must be valid")
	} else {
		u.IsValidUnit = true
	}

	// check that a regex pattern is passed
	if u.Pattern == "" {
		return errors.New("A valid regex pattern must be passed")
	}

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

				name := u.ReCompile.ReplaceAllString(key, u.Replacement)

				value, _ := toFloat(value)
				val, err := u.unitConverter(u.Unit, value, u.DestUnit)

				unitSuffix := strings.ToLower(u.DestUnit)

				if err != nil {
					u.Log.Errorf("Unable to convert unit: %v", err)

				} else {
					if !strings.HasSuffix(name, unitSuffix) {
						name += "_" + unitSuffix
					}

					if u.IsCounter {
						if !strings.HasSuffix(name, "_total") {
							name += "_total"
						}
					}

					metric.RemoveField(key)
					metric.AddField(name, val)
				}
			}
		}
	}

	return metrics
}

// unitConverter searches for the unit in
func (u *Units) unitConverter(unit string, v float64, d string) (float64, error) {

	var val interface{}

	if u.IsTemp {
		val, _ = convertTemperature(unit, v)
	} else {
		val, _ = u.AllUnits[unit]
	}

	value := reflect.ValueOf(val).MethodByName(d)
	if value.Kind() != 0 {
		convertedValue := value.Call([]reflect.Value{})[0].Float()
		if u.IsTemp {
			return convertedValue, nil
		} else {
			return v * convertedValue, nil
		}

	} else {
		return v, errors.New("Unit conversion method does not exist for the unit")
	}

}

// temperature is handled differently from thr unit package
func temperatureUnits() []string {
	units := []string{
		"Celsius",
		"Delisle",
		"Fahrenheit",
		"Kelvin",
		"Newton",
		"Rankine",
		"Reaumur",
		"Romer",
	}
	return units
}

// convertTemperature defines base temperature unit given a value
func convertTemperature(u string, v float64) (interface{}, error) {
	switch u {
	case "Celsius":
		return unit.FromCelsius(v), nil
	case "Delisle":
		return unit.FromDelisle(v), nil
	case "Fahrenheit":
		return unit.FromFahrenheit(v), nil
	case "Kelvin":
		return unit.FromKelvin(v), nil
	case "Newton":
		return unit.FromNewton(v), nil
	case "Rankine":
		return unit.FromRankine(v), nil
	case "Reaumur":
		return unit.FromReaumur(v), nil
	case "Romer":
		return unit.FromRomer(v), nil
	default:
		return nil, errors.New("Not a recognised unit of temperature")
	}
}

// allUnits creates a map of unit definitions
func allUnits() map[string]interface{} {
	units := map[string]interface{}{
		"Yoctoampere":                unit.Yoctoampere,
		"Zeptoampere":                unit.Zeptoampere,
		"Attoampere":                 unit.Attoampere,
		"Femtoampere":                unit.Femtoampere,
		"Picoampere":                 unit.Picoampere,
		"Nanoampere":                 unit.Nanoampere,
		"Microampere":                unit.Microampere,
		"Milliampere":                unit.Milliampere,
		"Deciampere":                 unit.Deciampere,
		"Centiampere":                unit.Centiampere,
		"Ampere":                     unit.Ampere,
		"Decaampere":                 unit.Decaampere,
		"Hectoampere":                unit.Hectoampere,
		"Kiloampere":                 unit.Kiloampere,
		"Megaampere":                 unit.Megaampere,
		"Gigaampere":                 unit.Gigaampere,
		"Teraampere":                 unit.Teraampere,
		"Petaampere":                 unit.Petaampere,
		"Exaampere":                  unit.Exaampere,
		"Zettaampere":                unit.Zettaampere,
		"Yottaampere":                unit.Yottaampere,
		"Newton":                     unit.Newton,
		"Dyne":                       unit.Dyne,
		"KilogramForce":              unit.KilogramForce,
		"PoundForce":                 unit.PoundForce,
		"Poundal":                    unit.Poundal,
		"Kilopond":                   unit.Kilopond,
		"Siemens":                    unit.Siemens,
		"Ohm":                        unit.Ohm,
		"Yoctopascal":                unit.Yoctopascal,
		"Zeptopascal":                unit.Zeptopascal,
		"Attopascal":                 unit.Attopascal,
		"Femtopascal":                unit.Femtopascal,
		"Picopascal":                 unit.Picopascal,
		"Nanopascal":                 unit.Nanopascal,
		"Micropascal":                unit.Micropascal,
		"Millipascal":                unit.Millipascal,
		"Centipascal":                unit.Centipascal,
		"Decipascal":                 unit.Decipascal,
		"Pascal":                     unit.Pascal,
		"Decapascal":                 unit.Decapascal,
		"Hectopascal":                unit.Hectopascal,
		"Kilopascal":                 unit.Kilopascal,
		"Megapascal":                 unit.Megapascal,
		"Gigapascal":                 unit.Gigapascal,
		"Terapascal":                 unit.Terapascal,
		"Petapascal":                 unit.Petapascal,
		"Exapascal":                  unit.Exapascal,
		"Zettapascal":                unit.Zettapascal,
		"Yottapascal":                unit.Yottapascal,
		"Yoctobar":                   unit.Yoctobar,
		"Zeptobar":                   unit.Zeptobar,
		"Attobar":                    unit.Attobar,
		"Femtobar":                   unit.Femtobar,
		"Picobar":                    unit.Picobar,
		"Nanobar":                    unit.Nanobar,
		"Microbar":                   unit.Microbar,
		"Millibar":                   unit.Millibar,
		"Centibar":                   unit.Centibar,
		"Decibar":                    unit.Decibar,
		"Bar":                        unit.Bar,
		"Decabar":                    unit.Decabar,
		"Hectobar":                   unit.Hectobar,
		"Kilobar":                    unit.Kilobar,
		"Megabar":                    unit.Megabar,
		"Gigabar":                    unit.Gigabar,
		"Terabar":                    unit.Terabar,
		"Petabar":                    unit.Petabar,
		"Exabar":                     unit.Exabar,
		"Zettabar":                   unit.Zettabar,
		"Yottabar":                   unit.Yottabar,
		"Atmosphere":                 unit.Atmosphere,
		"TechAtmosphere":             unit.TechAtmosphere,
		"Torr":                       unit.Torr,
		"PoundsPerSquareInch":        unit.PoundsPerSquareInch,
		"InchOfMercury":              unit.InchOfMercury,
		"Mole":                       unit.Mole,
		"Yoctogram":                  unit.Yoctogram,
		"Zeptogram":                  unit.Zeptogram,
		"Attogram":                   unit.Attogram,
		"Femtogram":                  unit.Femtogram,
		"Picogram":                   unit.Picogram,
		"Nanogram":                   unit.Nanogram,
		"Microgram":                  unit.Microgram,
		"Milligram":                  unit.Milligram,
		"Centigram":                  unit.Centigram,
		"Decigram":                   unit.Decigram,
		"Gram":                       unit.Gram,
		"Decagram":                   unit.Decagram,
		"Hectogram":                  unit.Hectogram,
		"Kilogram":                   unit.Kilogram,
		"Megagram":                   unit.Megagram,
		"Gigagram":                   unit.Gigagram,
		"Teragram":                   unit.Teragram,
		"Petagram":                   unit.Petagram,
		"Exagram":                    unit.Exagram,
		"Zettagram":                  unit.Zettagram,
		"Yottagram":                  unit.Yottagram,
		"Tonne":                      unit.Tonne,
		"Kilotonne":                  unit.Kilotonne,
		"Megatonne":                  unit.Megatonne,
		"Gigatonne":                  unit.Gigatonne,
		"Teratonne":                  unit.Teratonne,
		"Petatonne":                  unit.Petatonne,
		"Exatonne":                   unit.Exatonne,
		"TroyGrain":                  unit.TroyGrain,
		"AvoirdupoisDram":            unit.AvoirdupoisDram,
		"AvoirdupoisOunce":           unit.AvoirdupoisOunce,
		"AvoirdupoisPound":           unit.AvoirdupoisPound,
		"UsStone":                    unit.UsStone,
		"UsQuarter":                  unit.UsQuarter,
		"ShortHundredweight":         unit.ShortHundredweight,
		"UkStone":                    unit.UkStone,
		"UkQuarter":                  unit.UkQuarter,
		"LongHundredweight":          unit.LongHundredweight,
		"TroyOunce":                  unit.TroyOunce,
		"TroyPound":                  unit.TroyPound,
		"CentalHundredweight":        unit.CentalHundredweight,
		"ImperialHundredweight":      unit.ImperialHundredweight,
		"Yoctohertz":                 unit.Yoctohertz,
		"Zeptohertz":                 unit.Zeptohertz,
		"Attohertz":                  unit.Attohertz,
		"Femtohertz":                 unit.Femtohertz,
		"Picohertz":                  unit.Picohertz,
		"Nanohertz":                  unit.Nanohertz,
		"Microhertz":                 unit.Microhertz,
		"Millihertz":                 unit.Millihertz,
		"Centihertz":                 unit.Centihertz,
		"Decihertz":                  unit.Decihertz,
		"Hertz":                      unit.Hertz,
		"Decahertz":                  unit.Decahertz,
		"Hectohertz":                 unit.Hectohertz,
		"Kilohertz":                  unit.Kilohertz,
		"Megahertz":                  unit.Megahertz,
		"Gigahertz":                  unit.Gigahertz,
		"Terahertz":                  unit.Terahertz,
		"Petahertz":                  unit.Petahertz,
		"Exahertz":                   unit.Exahertz,
		"Zettahertz":                 unit.Zettahertz,
		"Yottahertz":                 unit.Yottahertz,
		"Lumen":                      unit.Lumen,
		"Yoctovolt":                  unit.Yoctovolt,
		"Zeptovolt":                  unit.Zeptovolt,
		"Attovolt":                   unit.Attovolt,
		"Femtovolt":                  unit.Femtovolt,
		"Picovolt":                   unit.Picovolt,
		"Nanovolt":                   unit.Nanovolt,
		"Microvolt":                  unit.Microvolt,
		"Millivolt":                  unit.Millivolt,
		"Centivolt":                  unit.Centivolt,
		"Decivolt":                   unit.Decivolt,
		"Volt":                       unit.Volt,
		"Decavolt":                   unit.Decavolt,
		"Hectovolt":                  unit.Hectovolt,
		"Kilovolt":                   unit.Kilovolt,
		"Megavolt":                   unit.Megavolt,
		"Gigavolt":                   unit.Gigavolt,
		"Teravolt":                   unit.Teravolt,
		"Petavolt":                   unit.Petavolt,
		"Exavolt":                    unit.Exavolt,
		"Zettavolt":                  unit.Zettavolt,
		"Yottavolt":                  unit.Yottavolt,
		"Kelvin":                     unit.Kelvin,
		"Yoctosecond":                unit.Yoctosecond,
		"Zeptosecond":                unit.Zeptosecond,
		"Attosecond":                 unit.Attosecond,
		"Femtosecond":                unit.Femtosecond,
		"Picosecond":                 unit.Picosecond,
		"Nanosecond":                 unit.Nanosecond,
		"Microsecond":                unit.Microsecond,
		"Millisecond":                unit.Millisecond,
		"Centisecond":                unit.Centisecond,
		"Decisecond":                 unit.Decisecond,
		"Second":                     unit.Second,
		"Decasecond":                 unit.Decasecond,
		"Hectosecond":                unit.Hectosecond,
		"Kilosecond":                 unit.Kilosecond,
		"Megasecond":                 unit.Megasecond,
		"Gigasecond":                 unit.Gigasecond,
		"Terasecond":                 unit.Terasecond,
		"Petasecond":                 unit.Petasecond,
		"Exasecond":                  unit.Exasecond,
		"Zettasecond":                unit.Zettasecond,
		"Yottasecond":                unit.Yottasecond,
		"Minute":                     unit.Minute,
		"Hour":                       unit.Hour,
		"Day":                        unit.Day,
		"Week":                       unit.Week,
		"ThirtyDayMonth":             unit.ThirtyDayMonth,
		"JulianYear":                 unit.JulianYear,
		"CentimeterPerSecondSquared": unit.CentimeterPerSecondSquared,
		"MeterPerSecondSquared":      unit.MeterPerSecondSquared,
		"FootPerSecondSquared":       unit.FootPerSecondSquared,
		"StandardGravity":            unit.StandardGravity,
		"Gal":                        unit.Gal,
		"Yoctoradian":                unit.Yoctoradian,
		"Zeptoradian":                unit.Zeptoradian,
		"Attoradian":                 unit.Attoradian,
		"Femtoradian":                unit.Femtoradian,
		"Picoradian":                 unit.Picoradian,
		"Nanoradian":                 unit.Nanoradian,
		"Microradian":                unit.Microradian,
		"Milliradian":                unit.Milliradian,
		"Centiradian":                unit.Centiradian,
		"Deciradian":                 unit.Deciradian,
		"Radian":                     unit.Radian,
		"Degree":                     unit.Degree,
		"Arcminute":                  unit.Arcminute,
		"Arcsecond":                  unit.Arcsecond,
		"Milliarcsecond":             unit.Milliarcsecond,
		"Microarcsecond":             unit.Microarcsecond,
		"Lux":                        unit.Lux,
		"Yoctometer":                 unit.Yoctometer,
		"Zeptometer":                 unit.Zeptometer,
		"Attometer":                  unit.Attometer,
		"Femtometer":                 unit.Femtometer,
		"Picometer":                  unit.Picometer,
		"Nanometer":                  unit.Nanometer,
		"Micrometer":                 unit.Micrometer,
		"Millimeter":                 unit.Millimeter,
		"Centimeter":                 unit.Centimeter,
		"Decimeter":                  unit.Decimeter,
		"Meter":                      unit.Meter,
		"Decameter":                  unit.Decameter,
		"Hectometer":                 unit.Hectometer,
		"Kilometer":                  unit.Kilometer,
		"ScandinavianMile":           unit.ScandinavianMile,
		"Megameter":                  unit.Megameter,
		"Gigameter":                  unit.Gigameter,
		"Terameter":                  unit.Terameter,
		"Petameter":                  unit.Petameter,
		"Exameter":                   unit.Exameter,
		"Zettameter":                 unit.Zettameter,
		"Yottameter":                 unit.Yottameter,
		"Inch":                       unit.Inch,
		"Hand":                       unit.Hand,
		"Foot":                       unit.Foot,
		"Yard":                       unit.Yard,
		"Link":                       unit.Link,
		"Rod":                        unit.Rod,
		"Chain":                      unit.Chain,
		"Furlong":                    unit.Furlong,
		"Mile":                       unit.Mile,
		"Fathom":                     unit.Fathom,
		"Cable":                      unit.Cable,
		"NauticalMile":               unit.NauticalMile,
		"LunarDistance":              unit.LunarDistance,
		"AstronomicalUnit":           unit.AstronomicalUnit,
		"LightYear":                  unit.LightYear,
		"MetersPerSecond":            unit.MetersPerSecond,
		"KilometersPerHour":          unit.KilometersPerHour,
		"FeetPerSecond":              unit.FeetPerSecond,
		"MilesPerHour":               unit.MilesPerHour,
		"Knot":                       unit.Knot,
		"SpeedOfLight":               unit.SpeedOfLight,
		"Yoctojoule":                 unit.Yoctojoule,
		"Zeptojoule":                 unit.Zeptojoule,
		"Attojoule":                  unit.Attojoule,
		"Femtojoule":                 unit.Femtojoule,
		"Picojoule":                  unit.Picojoule,
		"Nanojoule":                  unit.Nanojoule,
		"Microjoule":                 unit.Microjoule,
		"Millijoule":                 unit.Millijoule,
		"Centijoule":                 unit.Centijoule,
		"Decijoule":                  unit.Decijoule,
		"Joule":                      unit.Joule,
		"Decajoule":                  unit.Decajoule,
		"Hectojoule":                 unit.Hectojoule,
		"Kilojoule":                  unit.Kilojoule,
		"Megajoule":                  unit.Megajoule,
		"Gigajoule":                  unit.Gigajoule,
		"Terajoule":                  unit.Terajoule,
		"Petajoule":                  unit.Petajoule,
		"Exajoule":                   unit.Exajoule,
		"Zettajoule":                 unit.Zettajoule,
		"Yottajoule":                 unit.Yottajoule,
		"YoctowattHour":              unit.YoctowattHour,
		"ZeptowattHour":              unit.ZeptowattHour,
		"AttowattHour":               unit.AttowattHour,
		"FemtowattHour":              unit.FemtowattHour,
		"PicowattHour":               unit.PicowattHour,
		"NanowattHour":               unit.NanowattHour,
		"MicrowattHour":              unit.MicrowattHour,
		"MilliwattHour":              unit.MilliwattHour,
		"CentiwattHour":              unit.CentiwattHour,
		"DeciwattHour":               unit.DeciwattHour,
		"WattHour":                   unit.WattHour,
		"DecawattHour":               unit.DecawattHour,
		"HectowattHour":              unit.HectowattHour,
		"KilowattHour":               unit.KilowattHour,
		"MegawattHour":               unit.MegawattHour,
		"GigawattHour":               unit.GigawattHour,
		"TerawattHour":               unit.TerawattHour,
		"PetawattHour":               unit.PetawattHour,
		"ExawattHour":                unit.ExawattHour,
		"ZettawattHour":              unit.ZettawattHour,
		"YottawattHour":              unit.YottawattHour,
		"Gramcalorie":                unit.Gramcalorie,
		"Kilocalorie":                unit.Kilocalorie,
		"Megacalorie":                unit.Megacalorie,
		"CubicYoctometer":            unit.CubicYoctometer,
		"CubicZeptometer":            unit.CubicZeptometer,
		"CubicAttometer":             unit.CubicAttometer,
		"CubicFemtometer":            unit.CubicFemtometer,
		"CubicPicometer":             unit.CubicPicometer,
		"CubicNanometer":             unit.CubicNanometer,
		"CubicMicrometer":            unit.CubicMicrometer,
		"CubicMillimeter":            unit.CubicMillimeter,
		"CubicCentimeter":            unit.CubicCentimeter,
		"CubicDecimeter":             unit.CubicDecimeter,
		"CubicMeter":                 unit.CubicMeter,
		"CubicDecameter":             unit.CubicDecameter,
		"CubicHectometer":            unit.CubicHectometer,
		"CubicKilometer":             unit.CubicKilometer,
		"CubicMegameter":             unit.CubicMegameter,
		"CubicGigameter":             unit.CubicGigameter,
		"CubicTerameter":             unit.CubicTerameter,
		"CubicPetameter":             unit.CubicPetameter,
		"CubicExameter":              unit.CubicExameter,
		"CubicZettameter":            unit.CubicZettameter,
		"CubicYottameter":            unit.CubicYottameter,
		"Yoctoliter":                 unit.Yoctoliter,
		"Zepoliter":                  unit.Zepoliter,
		"Attoliter":                  unit.Attoliter,
		"Femtoliter":                 unit.Femtoliter,
		"Picoliter":                  unit.Picoliter,
		"Nanoliter":                  unit.Nanoliter,
		"Microliter":                 unit.Microliter,
		"Milliliter":                 unit.Milliliter,
		"Centiliter":                 unit.Centiliter,
		"Deciliter":                  unit.Deciliter,
		"Liter":                      unit.Liter,
		"Decaliter":                  unit.Decaliter,
		"Hectoliter":                 unit.Hectoliter,
		"Kiloliter":                  unit.Kiloliter,
		"Megaliter":                  unit.Megaliter,
		"Gigaliter":                  unit.Gigaliter,
		"Teraliter":                  unit.Teraliter,
		"Petaliter":                  unit.Petaliter,
		"Exaliter":                   unit.Exaliter,
		"Zettaliter":                 unit.Zettaliter,
		"Yottaliter":                 unit.Yottaliter,
		"CubicInch":                  unit.CubicInch,
		"CubicFoot":                  unit.CubicFoot,
		"CubicYard":                  unit.CubicYard,
		"CubicMile":                  unit.CubicMile,
		"CubicFurlong":               unit.CubicFurlong,
		"ImperialGallon":             unit.ImperialGallon,
		"ImperialQuart":              unit.ImperialQuart,
		"ImperialPint":               unit.ImperialPint,
		"ImperialCup":                unit.ImperialCup,
		"ImperialGill":               unit.ImperialGill,
		"ImperialFluidOunce":         unit.ImperialFluidOunce,
		"ImperialFluidDram":          unit.ImperialFluidDram,
		"ImperialPeck":               unit.ImperialPeck,
		"ImperialBushel":             unit.ImperialBushel,
		"MetricTableSpoon":           unit.MetricTableSpoon,
		"MetricTeaSpoon":             unit.MetricTeaSpoon,
		"USLiquidGallon":             unit.USLiquidGallon,
		"USLiquidQuart":              unit.USLiquidQuart,
		"USLiquidPint":               unit.USLiquidPint,
		"USCup":                      unit.USCup,
		"USLegalCup":                 unit.USLegalCup,
		"USGill":                     unit.USGill,
		"USFluidDram":                unit.USFluidDram,
		"USFluidOunce":               unit.USFluidOunce,
		"USTableSpoon":               unit.USTableSpoon,
		"USTeaSpoon":                 unit.USTeaSpoon,
		"USDryQuart":                 unit.USDryQuart,
		"USBushel":                   unit.USBushel,
		"USPeck":                     unit.USPeck,
		"USDryGallon":                unit.USDryGallon,
		"USDryPint":                  unit.USDryPint,
		"AustralianTableSpoon":       unit.AustralianTableSpoon,
		"ImperialTableSpoon":         unit.ImperialTableSpoon,
		"ImperialTeaSpoon":           unit.ImperialTeaSpoon,
		"BitPerSecond":               unit.BitPerSecond,
		"KilobitPerSecond":           unit.KilobitPerSecond,
		"MegabitPerSecond":           unit.MegabitPerSecond,
		"GigabitPerSecond":           unit.GigabitPerSecond,
		"TerabitPerSecond":           unit.TerabitPerSecond,
		"PetabitPerSecond":           unit.PetabitPerSecond,
		"ExabitPerSecond":            unit.ExabitPerSecond,
		"ZettabitPerSecond":          unit.ZettabitPerSecond,
		"YottabitPerSecond":          unit.YottabitPerSecond,
		"BytePerSecond":              unit.BytePerSecond,
		"KilobytePerSecond":          unit.KilobytePerSecond,
		"MegabytePerSecond":          unit.MegabytePerSecond,
		"GigabytePerSecond":          unit.GigabytePerSecond,
		"TerabytePerSecond":          unit.TerabytePerSecond,
		"PetabytePerSecond":          unit.PetabytePerSecond,
		"ExabytePerSecond":           unit.ExabytePerSecond,
		"ZettabytePerSecond":         unit.ZettabytePerSecond,
		"YottabytePerSecond":         unit.YottabytePerSecond,
		"KibibitPerSecond":           unit.KibibitPerSecond,
		"MebibitPerSecond":           unit.MebibitPerSecond,
		"GibibitPerSecond":           unit.GibibitPerSecond,
		"TebibitPerSecond":           unit.TebibitPerSecond,
		"PebibitPerSecond":           unit.PebibitPerSecond,
		"ExbibitPerSecond":           unit.ExbibitPerSecond,
		"ZebibitPerSecond":           unit.ZebibitPerSecond,
		"YobibitPerSecond":           unit.YobibitPerSecond,
		"KibibytePerSecond":          unit.KibibytePerSecond,
		"MebibytePerSecond":          unit.MebibytePerSecond,
		"GibibytePerSecond":          unit.GibibytePerSecond,
		"TebibytePerSecond":          unit.TebibytePerSecond,
		"PebibytePerSecond":          unit.PebibytePerSecond,
		"ExbibytePerSecond":          unit.ExbibytePerSecond,
		"ZebibytePerSecond":          unit.ZebibytePerSecond,
		"YobibytePerSecond":          unit.YobibytePerSecond,
		"Candela":                    unit.Candela,
		"Yoctowatt":                  unit.Yoctowatt,
		"Zeptowatt":                  unit.Zeptowatt,
		"Attowatt":                   unit.Attowatt,
		"Femtowatt":                  unit.Femtowatt,
		"Picowatt":                   unit.Picowatt,
		"Nanowatt":                   unit.Nanowatt,
		"Microwatt":                  unit.Microwatt,
		"Milliwatt":                  unit.Milliwatt,
		"Centiwatt":                  unit.Centiwatt,
		"Deciwatt":                   unit.Deciwatt,
		"Watt":                       unit.Watt,
		"Decawatt":                   unit.Decawatt,
		"Hectowatt":                  unit.Hectowatt,
		"Kilowatt":                   unit.Kilowatt,
		"Megawatt":                   unit.Megawatt,
		"Gigawatt":                   unit.Gigawatt,
		"Terawatt":                   unit.Terawatt,
		"Petawatt":                   unit.Petawatt,
		"Exawatt":                    unit.Exawatt,
		"Zettawatt":                  unit.Zettawatt,
		"Yottawatt":                  unit.Yottawatt,
		"Pferdestarke":               unit.Pferdestarke,
		"SquareYoctometer":           unit.SquareYoctometer,
		"SquareZeptometer":           unit.SquareZeptometer,
		"SquareAttometer":            unit.SquareAttometer,
		"SquareFemtometer":           unit.SquareFemtometer,
		"SquarePicometer":            unit.SquarePicometer,
		"SquareNanometer":            unit.SquareNanometer,
		"SquareMicrometer":           unit.SquareMicrometer,
		"SquareMillimeter":           unit.SquareMillimeter,
		"SquareCentimeter":           unit.SquareCentimeter,
		"SquareDecimeter":            unit.SquareDecimeter,
		"SquareMeter":                unit.SquareMeter,
		"SquareDecameter":            unit.SquareDecameter,
		"SquareHectometer":           unit.SquareHectometer,
		"SquareKilometer":            unit.SquareKilometer,
		"SquareMegameter":            unit.SquareMegameter,
		"SquareGigameter":            unit.SquareGigameter,
		"SquareTerameter":            unit.SquareTerameter,
		"SquarePetameter":            unit.SquarePetameter,
		"SquareExameter":             unit.SquareExameter,
		"SquareZettameter":           unit.SquareZettameter,
		"SquareYottameter":           unit.SquareYottameter,
		"SquareInch":                 unit.SquareInch,
		"SquareFoot":                 unit.SquareFoot,
		"SquareYard":                 unit.SquareYard,
		"Acre":                       unit.Acre,
		"SquareMile":                 unit.SquareMile,
		"TWGongQing":                 unit.TWGongQing,
		"TWPing":                     unit.TWPing,
		"TWFen":                      unit.TWFen,
		"TWJia":                      unit.TWJia,
		"SquareRod":                  unit.SquareRod,
		"Rood":                       unit.Rood,
		"Centiare":                   unit.Centiare,
		"Are":                        unit.Are,
		"Hectare":                    unit.Hectare,
		"SquarePerch":                unit.SquarePerch,
		"Bit":                        unit.Bit,
		"Kilobit":                    unit.Kilobit,
		"Megabit":                    unit.Megabit,
		"Gigabit":                    unit.Gigabit,
		"Terabit":                    unit.Terabit,
		"Petabit":                    unit.Petabit,
		"Exabit":                     unit.Exabit,
		"Zettabit":                   unit.Zettabit,
		"Yottabit":                   unit.Yottabit,
		"Byte":                       unit.Byte,
		"Kilobyte":                   unit.Kilobyte,
		"Megabyte":                   unit.Megabyte,
		"Gigabyte":                   unit.Gigabyte,
		"Terabyte":                   unit.Terabyte,
		"Petabyte":                   unit.Petabyte,
		"Exabyte":                    unit.Exabyte,
		"Zettabyte":                  unit.Zettabyte,
		"Yottabyte":                  unit.Yottabyte,
		"Kibibit":                    unit.Kibibit,
		"Mebibit":                    unit.Mebibit,
		"Gibibit":                    unit.Gibibit,
		"Tebibit":                    unit.Tebibit,
		"Pebibit":                    unit.Pebibit,
		"Exbibit":                    unit.Exbibit,
		"Zebibit":                    unit.Zebibit,
		"Yobibit":                    unit.Yobibit,
		"Kibibyte":                   unit.Kibibyte,
		"Mebibyte":                   unit.Mebibyte,
		"Gibibyte":                   unit.Gibibyte,
		"Tebibyte":                   unit.Tebibyte,
		"Pebibyte":                   unit.Pebibyte,
		"Exbibyte":                   unit.Exbibyte,
		"Zebibyte":                   unit.Zebibyte,
		"Yobibyte":                   unit.Yobibyte,
	}
	return units
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
