/******************************************************************************/
/* shader_graph_node_catalog.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import (
	"strings"

	"kaijuengine.com/matrix"
)

type shaderGraphNodeCatalogEntry struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Spec        shaderGraphNodeSpec
}

type shaderGraphNodeMenuData struct {
	ID          string
	Name        string
	Description string
	Search      string
}

func shaderGraphNodeCatalog() []shaderGraphNodeCatalogEntry {
	return []shaderGraphNodeCatalogEntry{
		{
			ID:          "value",
			Name:        "Value",
			Description: "Single float value.",
			Tags:        []string{"float", "number", "constant"},
			Spec: shaderGraphNodeSpec{
				Name:        "Value",
				Description: "Single float value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:      "value",
						Label:   "Value",
						Type:    shaderGraphNodeFieldNumber,
						Default: "1.000",
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Value", Type: "float"},
				},
			},
		},
		{
			ID:          "color",
			Name:        "Color",
			Description: "Single color value.",
			Tags:        []string{"color", "constant", "albedo"},
			Spec: shaderGraphNodeSpec{
				Name:        "Color",
				Description: "Single color value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:           "color",
						Label:        "Color",
						Type:         shaderGraphNodeFieldColor,
						DefaultColor: matrix.ColorWhite(),
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "vector",
			Name:        "Vector",
			Description: "Single vec3 value.",
			Tags:        []string{"vector", "vec3", "constant"},
			Spec: shaderGraphNodeSpec{
				Name:        "Vector",
				Description: "Single vec3 value.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:            "vector",
						Label:         "Vector",
						Type:          shaderGraphNodeFieldVector3,
						DefaultValues: []string{"0", "0", "0"},
					},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Vector", Type: "vec3"},
				},
			},
		},
		{
			ID:          "mix-color",
			Name:        "Mix Color",
			Description: "Blends two colors with a factor.",
			Tags:        []string{"mix", "blend", "color", "factor"},
			Spec: shaderGraphNodeSpec{
				Name:        "Mix Color",
				Description: "Blends two colors with a factor.",
				Fields: []shaderGraphNodeFieldSpec{
					{
						ID:          "clamp",
						Label:       "Clamp",
						Type:        shaderGraphNodeFieldBool,
						DefaultBool: true,
					},
					{
						ID:      "mode",
						Label:   "Mode",
						Type:    shaderGraphNodeFieldSelect,
						Default: "mix",
						Options: []shaderGraphNodeFieldOption{
							{Label: "Mix", Value: "mix"},
							{Label: "Add", Value: "add"},
							{Label: "Multiply", Value: "multiply"},
						},
					},
				},
				Inputs: []shaderGraphPortSpec{
					{Name: "Factor", Type: "float"},
					{Name: "A", Type: "color"},
					{Name: "B", Type: "color"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "Color", Type: "color"},
				},
			},
		},
		{
			ID:          "principled-bsdf",
			Name:        "Principled BSDF",
			Description: "Surface shader with common material inputs.",
			Tags:        []string{"bsdf", "surface", "material", "shader"},
			Spec: shaderGraphNodeSpec{
				Name:        "Principled BSDF",
				Description: "Surface shader with common material inputs.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Base Color", Type: "color"},
					{Name: "Roughness", Type: "float"},
					{Name: "Normal", Type: "vec3"},
				},
				Outputs: []shaderGraphPortSpec{
					{Name: "BSDF", Type: "surface"},
				},
			},
		},
		{
			ID:          "material-output",
			Name:        "Material Output",
			Description: "Terminal output for the material shader.",
			Tags:        []string{"output", "surface", "volume", "material"},
			Spec: shaderGraphNodeSpec{
				Name:        "Material Output",
				Description: "Terminal output for the material shader.",
				Inputs: []shaderGraphPortSpec{
					{Name: "Surface", Type: "surface"},
					{Name: "Volume", Type: "volume"},
					{Name: "Displacement", Type: "vec3"},
				},
			},
		},
	}
}

func shaderGraphNodeCatalogMenuData() []shaderGraphNodeMenuData {
	catalog := shaderGraphNodeCatalog()
	data := make([]shaderGraphNodeMenuData, 0, len(catalog))
	for i := range catalog {
		entry := catalog[i]
		search := strings.Join(append([]string{entry.ID, entry.Name, entry.Description}, entry.Tags...), " ")
		data = append(data, shaderGraphNodeMenuData{
			ID:          entry.ID,
			Name:        entry.Name,
			Description: entry.Description,
			Search:      strings.ToLower(search),
		})
	}
	return data
}

func shaderGraphNodeCatalogSpec(id string) (shaderGraphNodeSpec, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, entry := range shaderGraphNodeCatalog() {
		if entry.ID == id {
			return entry.Spec, true
		}
	}
	return shaderGraphNodeSpec{}, false
}
