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

// PrincipalType the model 'PrincipalType'
type PrincipalType string

// List of PrincipalType
const (
	LU PrincipalType = "LU"
	LG PrincipalType = "LG"
	DU PrincipalType = "DU"
	DG PrincipalType = "DG"
)

var allowedPrincipalTypeEnumValues = []PrincipalType{
	"LU",
	"LG",
	"DU",
	"DG",
}

func (v *PrincipalType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := PrincipalType(value)
	for _, existing := range allowedPrincipalTypeEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid PrincipalType", value)
}

// NewPrincipalTypeFromValue returns a pointer to a valid PrincipalType
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewPrincipalTypeFromValue(v string) (*PrincipalType, error) {
	ev := PrincipalType(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for PrincipalType: valid values are %v", v, allowedPrincipalTypeEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v PrincipalType) IsValid() bool {
	for _, existing := range allowedPrincipalTypeEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to PrincipalType value
func (v PrincipalType) Ptr() *PrincipalType {
	return &v
}

type NullablePrincipalType struct {
	value *PrincipalType
	isSet bool
}

func (v NullablePrincipalType) Get() *PrincipalType {
	return v.value
}

func (v *NullablePrincipalType) Set(val *PrincipalType) {
	v.value = val
	v.isSet = true
}

func (v NullablePrincipalType) IsSet() bool {
	return v.isSet
}

func (v *NullablePrincipalType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePrincipalType(val *PrincipalType) *NullablePrincipalType {
	return &NullablePrincipalType{value: val, isSet: true}
}

func (v NullablePrincipalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePrincipalType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
