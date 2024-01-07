package config

import (
	"io"
	"log"

	"github.com/samlitowitz/goimportcycle/internal/color"
)

type Resolution int

const (
	FileResolution    Resolution = iota
	PackageResolution Resolution = iota
)

type Config struct {
	Palette *color.Palette

	Resolution Resolution
	Debug      *log.Logger
}

func Default() *Config {
	return &Config{
		Palette:    color.Default,
		Resolution: FileResolution,
		Debug:      log.New(io.Discard, "Debug: ", log.LstdFlags),
	}
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
