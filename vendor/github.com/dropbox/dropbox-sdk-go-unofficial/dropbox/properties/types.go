// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// This namespace contains helper entities for property and property/template
// endpoints.
package properties

import (
	"encoding/json"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
)

type GetPropertyTemplateArg struct {
	// An identifier for property template added by route
	// properties/template/add.
	TemplateId string `json:"template_id"`
}

func NewGetPropertyTemplateArg(TemplateId string) *GetPropertyTemplateArg {
	s := new(GetPropertyTemplateArg)
	s.TemplateId = TemplateId
	return s
}

// Describes property templates that can be filled and associated with a file.
type PropertyGroupTemplate struct {
	// A display name for the property template. Property template names can be
	// up to 256 bytes.
	Name string `json:"name"`
	// Description for new property template. Property template descriptions can
	// be up to 1024 bytes.
	Description string `json:"description"`
	// This is a list of custom properties associated with a property template.
	// There can be up to 64 properties in a single property template.
	Fields []*PropertyFieldTemplate `json:"fields"`
}

func NewPropertyGroupTemplate(Name string, Description string, Fields []*PropertyFieldTemplate) *PropertyGroupTemplate {
	s := new(PropertyGroupTemplate)
	s.Name = Name
	s.Description = Description
	s.Fields = Fields
	return s
}

// The Property template for the specified template.
type GetPropertyTemplateResult struct {
	PropertyGroupTemplate
}

func NewGetPropertyTemplateResult(Name string, Description string, Fields []*PropertyFieldTemplate) *GetPropertyTemplateResult {
	s := new(GetPropertyTemplateResult)
	s.Name = Name
	s.Description = Description
	s.Fields = Fields
	return s
}

type ListPropertyTemplateIds struct {
	// List of identifiers for templates added by route properties/template/add.
	TemplateIds []string `json:"template_ids"`
}

func NewListPropertyTemplateIds(TemplateIds []string) *ListPropertyTemplateIds {
	s := new(ListPropertyTemplateIds)
	s.TemplateIds = TemplateIds
	return s
}

type PropertyTemplateError struct {
	dropbox.Tagged
	// Property template does not exist for given identifier.
	TemplateNotFound string `json:"template_not_found,omitempty"`
}

const (
	PropertyTemplateError_TemplateNotFound  = "template_not_found"
	PropertyTemplateError_RestrictedContent = "restricted_content"
	PropertyTemplateError_Other             = "other"
)

func (u *PropertyTemplateError) UnmarshalJSON(body []byte) error {
	type wrap struct {
		dropbox.Tagged
	}
	var w wrap
	if err := json.Unmarshal(body, &w); err != nil {
		return err
	}
	u.Tag = w.Tag
	switch u.Tag {
	case "template_not_found":
		if err := json.Unmarshal(body, &u.TemplateNotFound); err != nil {
			return err
		}

	}
	return nil
}

type ModifyPropertyTemplateError struct {
	dropbox.Tagged
}

const (
	ModifyPropertyTemplateError_ConflictingPropertyNames  = "conflicting_property_names"
	ModifyPropertyTemplateError_TooManyProperties         = "too_many_properties"
	ModifyPropertyTemplateError_TooManyTemplates          = "too_many_templates"
	ModifyPropertyTemplateError_TemplateAttributeTooLarge = "template_attribute_too_large"
)

type PropertyField struct {
	// This is the name or key of a custom property in a property template. File
	// property names can be up to 256 bytes.
	Name string `json:"name"`
	// Value of a custom property attached to a file. Values can be up to 1024
	// bytes.
	Value string `json:"value"`
}

func NewPropertyField(Name string, Value string) *PropertyField {
	s := new(PropertyField)
	s.Name = Name
	s.Value = Value
	return s
}

// Describe a single property field type which that can be part of a property
// template.
type PropertyFieldTemplate struct {
	// This is the name or key of a custom property in a property template. File
	// property names can be up to 256 bytes.
	Name string `json:"name"`
	// This is the description for a custom property in a property template.
	// File property description can be up to 1024 bytes.
	Description string `json:"description"`
	// This is the data type of the value of this property. This type will be
	// enforced upon property creation and modifications.
	Type *PropertyType `json:"type"`
}

func NewPropertyFieldTemplate(Name string, Description string, Type *PropertyType) *PropertyFieldTemplate {
	s := new(PropertyFieldTemplate)
	s.Name = Name
	s.Description = Description
	s.Type = Type
	return s
}

// Collection of custom properties in filled property templates.
type PropertyGroup struct {
	// A unique identifier for a property template type.
	TemplateId string `json:"template_id"`
	// This is a list of custom properties associated with a file. There can be
	// up to 32 properties for a template.
	Fields []*PropertyField `json:"fields"`
}

func NewPropertyGroup(TemplateId string, Fields []*PropertyField) *PropertyGroup {
	s := new(PropertyGroup)
	s.TemplateId = TemplateId
	s.Fields = Fields
	return s
}

// Data type of the given property added. This endpoint is in beta and  only
// properties of type strings is supported.
type PropertyType struct {
	dropbox.Tagged
}

const (
	PropertyType_String = "string"
	PropertyType_Other  = "other"
)
