package config

import (
	"fmt"
	stdColor "image/color"
	"io"
	"log"
	"os"

	"github.com/go-playground/colors"
	inColor "github.com/samlitowitz/goimportcycle/internal/color"
	"gopkg.in/yaml.v3"
)

type Resolution int

const (
	FileResolution    Resolution = iota
	PackageResolution Resolution = iota
)

type Config struct {
	Palette *inColor.Palette

	Resolution Resolution
	Debug      *log.Logger
}

type externalPalette struct {
	PackageName       string `yaml:"packageName,omitempty"`
	PackageBackground string `yaml:"packageBackground,omitempty"`
	FileName          string `yaml:"fileName,omitempty"`
	FileBackground    string `yaml:"fileBackground,omitempty"`
	ImportArrow       string `yaml:"importArrow,omitempty"`
}

type externalConfig struct {
	Resolution string `yaml:"resolution,omitempty"`
	Palette    *struct {
		Base  *externalPalette `yaml:"base,omitempty"`
		Cycle *externalPalette `yaml:"cycle,omitempty"`
	} `yaml:"palette,omitempty"`
}

func Default() *Config {
	return &Config{
		Palette:    inColor.Default,
		Resolution: FileResolution,
		Debug:      log.New(io.Discard, "Debug: ", log.LstdFlags),
	}
}

func FromYamlFile(filepath string) (*Config, error) {
	f, err := os.Open(filepath)

	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ecfg *externalConfig
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(ecfg)
	if err == io.EOF {
		return Default(), nil
	}
	if err != nil {
		return nil, err
	}
	cfg := Default()
	err = fromExternalConfig(cfg, ecfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func fromExternalConfig(to *Config, from *externalConfig) error {
	// resolution
	switch from.Resolution {
	case "":
	case "file":
		to.Resolution = FileResolution
	case "package":
		to.Resolution = PackageResolution

	default:
		return fmt.Errorf(
			"invalid resolution %s, must be one of ['file', 'package']",
			from.Resolution,
		)
	}

	// palette
	if from.Palette == nil {
		return nil
	}

	// base palette
	err := fromExternalPalette(to.Palette.Base, from.Palette.Base)
	if err != nil {
		return err
	}

	// cycle palette
	err = fromExternalPalette(to.Palette.Cycle, from.Palette.Cycle)
	if err != nil {
		return err
	}

	return nil
}

func fromExternalPalette(to *inColor.HalfPalette, from *externalPalette) error {
	if from == nil {
		return nil
	}

	tryParseColor := func(in string) (stdColor.Color, error) {
		if in == "" {
			return nil, nil
		}
		hex, err := colors.ParseHEX(in)
		if err == nil {
			return hex, nil
		}
		rgb, err := colors.ParseRGB(in)
		if err == nil {
			return rgb, nil
		}
		rgba, err := colors.ParseRGBA(in)
		if err == nil {
			return rgba, nil
		}

		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to parse color %s", in)
	}

	if from.PackageName != "" {
		c, err := tryParseColor(from.PackageName)
		if err != nil {
			return err
		}
		if c != nil {
			to.PackageName = inColor.Color{Color: c}
		}
	}
	if from.PackageBackground != "" {
		c, err := tryParseColor(from.PackageBackground)
		if err != nil {
			return err
		}
		if c != nil {
			to.PackageBackground = inColor.Color{Color: c}
		}
	}
	if from.FileName != "" {
		c, err := tryParseColor(from.FileName)
		if err != nil {
			return err
		}
		if c != nil {
			to.FileName = inColor.Color{Color: c}
		}
	}
	if from.FileBackground != "" {
		c, err := tryParseColor(from.FileBackground)
		if err != nil {
			return err
		}
		if c != nil {
			to.FileBackground = inColor.Color{Color: c}
		}
	}
	if from.ImportArrow != "" {
		c, err := tryParseColor(from.ImportArrow)
		if err != nil {
			return err
		}
		if c != nil {
			to.ImportArrow = inColor.Color{Color: c}
		}
	}
	return nil
}

//# File Resolution
//1. package name
//1. package background
//1. file name
//1. file background
//1. import arrow
//
//1. not in cycle/in cycle
//
//
//# Package Resolution
//1. package name
//1. package background
//1. import arrow
//
//1. not in cycle/in cycle
