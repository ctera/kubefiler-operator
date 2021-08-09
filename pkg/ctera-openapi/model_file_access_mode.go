/*
 * CTERA Gateway
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package cteraopenapi

import (
	"encoding/json"
	"fmt"
)

// FileAccessMode the model 'FileAccessMode'
type FileAccessMode string

// List of FileAccessMode
const (
	RW FileAccessMode = "RW"
	RO FileAccessMode = "RO"
	NA FileAccessMode = "NA"
)

var allowedFileAccessModeEnumValues = []FileAccessMode{
	"RW",
	"RO",
	"NA",
}

func (v *FileAccessMode) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := FileAccessMode(value)
	for _, existing := range allowedFileAccessModeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid FileAccessMode", value)
}

// NewFileAccessModeFromValue returns a pointer to a valid FileAccessMode
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewFileAccessModeFromValue(v string) (*FileAccessMode, error) {
	ev := FileAccessMode(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for FileAccessMode: valid values are %v", v, allowedFileAccessModeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v FileAccessMode) IsValid() bool {
	for _, existing := range allowedFileAccessModeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to FileAccessMode value
func (v FileAccessMode) Ptr() *FileAccessMode {
	return &v
}

type NullableFileAccessMode struct {
	value *FileAccessMode
	isSet bool
}

func (v NullableFileAccessMode) Get() *FileAccessMode {
	return v.value
}

func (v *NullableFileAccessMode) Set(val *FileAccessMode) {
	v.value = val
	v.isSet = true
}

func (v NullableFileAccessMode) IsSet() bool {
	return v.isSet
}

func (v *NullableFileAccessMode) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFileAccessMode(val *FileAccessMode) *NullableFileAccessMode {
	return &NullableFileAccessMode{value: val, isSet: true}
}

func (v NullableFileAccessMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFileAccessMode) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}