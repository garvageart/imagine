package settings

import (
	"encoding/json"
	"fmt"
	"strconv"

	"imagine/internal/dto"
	"imagine/internal/entities"
)

// Helper function for boolean settings
func BoolSetting(name string, value bool, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		Value:          strconv.FormatBool(value),
		ValueType:      dto.Boolean, 
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for integer settings
func IntSetting(name string, value int, allowedValues []int, isUserEditable bool, group, description string) entities.SettingDefault {
	var allowedValuesStr *[]string
	if len(allowedValues) > 0 {
		strValues := make([]string, len(allowedValues))
		for i, v := range allowedValues {
			strValues[i] = strconv.Itoa(v)
		}
		allowedValuesStr = &strValues
	}

	return entities.SettingDefault{
		Name:           name,
		Value:          strconv.Itoa(value),
		ValueType:      dto.Integer, 
		AllowedValues:  allowedValuesStr,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for string settings
func StringSetting(name string, value string, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		Value:          value,
		ValueType:      dto.String, 
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for enum settings
func EnumSetting(name string, value string, allowedValues []string, isUserEditable bool, group, description string) entities.SettingDefault {
	return entities.SettingDefault{
		Name:           name,
		Value:          value,
		ValueType:      dto.Enum, 
		AllowedValues:  &allowedValues,
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}

// Helper function for JSON settings
func JsonSetting(name string, value interface{}, isUserEditable bool, group, description string) entities.SettingDefault {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON for setting %s: %v", name, err))
	}

	return entities.SettingDefault{
		Name:           name,
		Value:          string(jsonBytes),
		ValueType:      dto.Json, 
		IsUserEditable: isUserEditable,
		Group:          group,
		Description:    description,
	}
}