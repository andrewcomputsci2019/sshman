package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type optionFlags []string

func (o *optionFlags) String() string {
	return fmt.Sprintf("%v", *o)
}

func (o *optionFlags) Set(value string) error {
	*o = append(*o, value)
	return nil
}

type SettableFlag[T any] struct {
	Value     T
	SetByUser bool
	conv      func(source string) (T, error)
}

func (s *SettableFlag[T]) String() string {
	if s == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", s.Value)
}

func (s *SettableFlag[T]) Set(val string) error {
	v, err := s.conv(val)
	if err != nil {
		return err
	}
	s.Value = v
	s.SetByUser = true
	return nil
}

func NewSettableFlag[T any](name string, defaultValue T, usage string, conv func(s string) (T, error)) *SettableFlag[T] {
	f := &SettableFlag[T]{
		Value:     defaultValue,
		SetByUser: false,
		conv:      conv,
	}
	flag.Var(f, name, usage)
	return f
}

func NewStringSettableFlag(name string, defaultValue string, usage string) *SettableFlag[string] {
	return NewSettableFlag(name, defaultValue, usage, func(s string) (string, error) {
		return s, nil
	})
}

func NewUintSettableFlag(name string, defaultValue uint, usage string) *SettableFlag[uint] {
	return NewSettableFlag(name, defaultValue, usage, func(s string) (uint, error) {
		if s == "" {
			return 0, nil
		}
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(v), nil
	})
}

func NewIntSettableFlag(name string, defaultValue int, usage string) *SettableFlag[int] {
	return NewSettableFlag(name, defaultValue, usage, func(s string) (int, error) {
		if s == "" {
			return 0, nil
		}
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return 0, err
		}
		return int(v), nil
	})
}

func main() {
	logLevel := flag.String("log", "None", "logging level, [Debug, Info, Warning, Error]")
	quickAdd := flag.Bool("qa", false, "quick add")
	quickDelete := flag.Bool("qd", false, "quick delete")
	quickEdit := flag.Bool("qe", false, "quick edit")
	quickConnect := flag.Bool("qc", false, "quick connect")
	host := flag.String("h", "", "host")
	port := flag.Uint("p", 22, "port")
	identityFile := flag.String("i", "", "identity file, used in qc and qe commands")
	var optionFlags optionFlags
	flag.Var(&optionFlags, "o",
		"option flags, these are ssh options that you want passed, works for quick commands such as qe, and qa, in tui, "+
			"these are passed when invoking ssh, these should be passed identically to how they are written in a config file")
	flag.Parse()
	if logLevel != nil && *logLevel != "None" {
		// set up slog here
		lowerCase := strings.ToLower(*logLevel)
		switch lowerCase {
		case "debug":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		case "info":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		case "warning":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelWarn,
			})))
		case "error":
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelError,
			})))
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unkown log level: %s, defualting to info\n", *logLevel)
			slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})))
		}
	}

	// detect if quick commands are passed
	if *quickConnect || *quickEdit || *quickAdd || *quickDelete {

	}
}
