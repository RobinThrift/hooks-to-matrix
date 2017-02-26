package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path"
	"reflect"
	"strings"
)

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(strings.Title(name))

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return fmt.Errorf("Provided value %s type didn't match obj field type of %s", val.Type().String(), name)
	}

	if structFieldType.String() == "string" {
		structFieldValue.Set(reflect.ValueOf(ReplaceValueWithEnvVar(value.(string))))
	} else {
		structFieldValue.Set(val)
	}

	return nil
}

func ReplaceValueWithEnvVar(value string) string {
	if strings.Index(value, "$") != 0 {
		return value
	}

	return os.Getenv(strings.Replace(value, "$", "", 1))
}

// LoadConfig loads a config file from the given path
func LoadConfig(configPath string) (*viper.Viper, error) {
	config := viper.New()
	dir := path.Dir(configPath)
	extname := path.Ext(configPath)
	name := strings.Replace(path.Base(configPath), extname, "", 1)

	config.SetEnvPrefix("htm")
	config.AutomaticEnv()
	config.SetConfigName(name)
	config.AddConfigPath(dir)

	config.SetDefault("port", ":8080")

	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}

	return config, nil
}
