package schema

import (
	"fmt"
	"reflect"
)

type Schema struct {
	Overrides map[string]func(t reflect.Type)
	Json      Json
}

func NewSchema(name string, version string) *Schema {
	return &Schema{
		Json: Json{
			Components:   []Component{},
			ConfigBlocks: []ConfigBlock{},
			Services:     []Service{},
			Name:         name,
			Version:      version,
		},
	}
}

func (s *Schema) AddComponent(name string, description string, args any, exports any) error {
	c := Component{
		Arguments:   make([]Type, 0),
		Exports:     make([]Type, 0),
		Name:        name,
		Description: description,
	}
	argTypes, err := parseTypes(reflect.TypeOf(args))
	if err != nil {
		return err
	}
	c.Arguments = argTypes
	exportTypes, err := parseTypes(reflect.TypeOf(exports))
	if err != nil {
		return err
	}
	c.Exports = exportTypes
	s.Json.Components = append(s.Json.Components, c)
	return nil
}

func (s *Schema) AddService(name string, description string, args any) error {
	service := Service{
		Name:        name,
		Description: description,
	}
	argTypes, err := parseTypes(reflect.TypeOf(args))
	if err != nil {
		return err
	}
	service.Arguments = argTypes
	s.Json.Services = append(s.Json.Services, service)
	return nil
}

func (s *Schema) AddConfigBlock(name string, description string, args any) error {
	config := ConfigBlock{
		Name:        name,
		Description: description,
	}
	argTypes, err := parseTypes(reflect.TypeOf(args))
	if err != nil {
		return err
	}
	config.Arguments = argTypes
	s.Json.ConfigBlocks = append(s.Json.ConfigBlocks, config)
	return nil
}

func parseTypes(t reflect.Type) ([]Type, error) {
	results := make([]Type, 0)
	if t == nil {
		return results, nil
	}
	tags := Get(t)
	for _, t := range tags {
		st := Type{
			Children: make([]Type, 0),
			Args:     make([]Arg, 0),
		}
		st.Name = t.Name[0]
		switch t.Type.Kind() {
		case reflect.String:
			st.Type = "string"
		case reflect.Bool:
			st.Type = "bool"
		case reflect.Int:
			st.Type = "int"
		case reflect.Struct:
			st.Type = "object"
			children, err := parseTypes(t.Type)
			if err != nil {
				return nil, err
			}
			st.Children = children
		case reflect.Slice:
			st.Type = "array"
			children, err := parseTypes(t.Type.Elem())
			if err != nil {
				return nil, err
			}
			st.Children = children
		case reflect.Map:
			st.Type = "map"
			children, err := parseTypes(t.Type.Elem())
			if err != nil {
				return nil, err
			}
			st.Children = children
		default:
			return nil, fmt.Errorf("unknown type %d", t.Type.Kind())
		}
		st.Optional = t.IsOptional()
		results = append(results, st)
	}
	return results, nil
}

type Json struct {
	Components   []Component   `json:"components"`
	ConfigBlocks []ConfigBlock `json:"config_blocks"`
	Services     []Service     `json:"services"`
	Name         string        `json:"name"`
	Version      string        `json:"version"`
}

type Component struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   []Type `json:"arguments"`
	Exports     []Type `json:"Exports"`
}

type ConfigBlock struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   []Type `json:"arguments"`
}

type Service struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   []Type `json:"arguments"`
}

type Type struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Children    []Type `json:"children"`
	Args        []Arg  `json:"args"`
	Optional    bool   `json:"optional"`
}

type Arg struct {
	Name        string `json:"name"`
	Type        Type   `json:"type"`
	Description string `json:"description"`
}
