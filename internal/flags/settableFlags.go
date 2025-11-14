package flags

import (
	"flag"
	"fmt"
	"log/slog"
	"strconv"
)

type SettableFlag[T any] struct {
	Value     T
	flagName  string
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
	slog.Debug("[SettableFlag] setting flag", "flag", s.flagName, "value", val)
	v, err := s.conv(val)
	if err != nil {
		slog.Error("[SettableFlag] failed to convert value", "flag", s.flagName, "value", val)
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
		flagName:  name,
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
