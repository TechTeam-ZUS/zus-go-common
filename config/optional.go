package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// LoadOptional fills dst — a pointer to a consumer-defined struct — from
// environment variables, based on each field's `env:"KEY"` tag. Call it
// after Load() so .env values are already present in the process
// environment.
//
// The config package never needs to know what fields exist; each project
// defines its own struct, and LoadOptional just maps tagged fields onto it.
//
// Supported tag options (comma-separated after the key):
//
//	env:"KEY"                  // unset -> zero value
//	env:"KEY,required"         // unset -> error
//	env:"KEY,default=value"    // unset -> fallback value
//
// Supported field kinds: string, bool, int/int8/../int64,
// float32/float64, time.Duration, and []string (comma-separated).
// Fields without an `env` tag are skipped, so consumers can mix in
// unexported/helper fields freely.
//
// Example:
//
//	type OptionalConfig struct {
//	    FeatureFlagX   bool          `env:"FEATURE_FLAG_X"`
//	    MaxQueueSize   int           `env:"MAX_QUEUE_SIZE,default=100"`
//	    CacheTTL       time.Duration `env:"CACHE_TTL,default=5m"`
//	    PaymentWebhook string        `env:"PAYMENT_WEBHOOK_URL,required"`
//	}
//
//	var Optional OptionalConfig
//	if err := config.Load(); err != nil { ... }
//	if err := config.LoadOptional(&Optional); err != nil { ... }
//	Optional.MaxQueueSize // read directly, no re-parsing env

func LoadOptional(dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config: LoadOptional: dst must be a non-nil pointer to a struct, got %T", dst)
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup("env")
		if !ok || tag == "" {
			continue
		}

		key, required, def := parseOptionalTag(tag)
		raw := os.Getenv(key)
		if raw == "" {
			if required {
				return fmt.Errorf("config: LoadOptional: %s: required but not set", key)
			}
			raw = def
			if raw == "" {
				continue // leave field at zero value
			}
		}

		if err := setOptionalField(v.Field(i), raw); err != nil {
			return fmt.Errorf("config: LoadOptional: %s: %w", key, err)
		}
	}

	return nil
}

func parseOptionalTag(tag string) (key string, required bool, def string) {
	parts := strings.Split(tag, ",")
	key = parts[0]
	for _, opt := range parts[1:] {
		switch {
		case opt == "required":
			required = true
		case strings.HasPrefix(opt, "default="):
			def = strings.TrimPrefix(opt, "default=")
		}
	}
	return key, required, def
}

func setOptionalField(field reflect.Value, raw string) error {
	// time.Duration is an int64 under the hood, so it must be checked
	// before the generic Int case below.
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", raw, err)
		}
		field.SetInt(int64(d))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", raw, err)
		}
		field.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int %q: %w", raw, err)
		}
		field.SetInt(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", raw, err)
		}
		field.SetFloat(f)
	case reflect.Slice:
		if field.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("unsupported slice element type %s", field.Type().Elem())
		}
		parts := strings.Split(raw, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		field.Set(reflect.ValueOf(parts))
	default:
		return fmt.Errorf("unsupported field type %s", field.Type())
	}
	return nil
}
