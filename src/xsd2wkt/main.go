package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"
)

// XSD structure to hold parsed data
type XSD struct {
	Elements []Element `xml:"element"`
}

// Add Type field to Element struct
type Element struct {
	Name     string    `xml:"name,attr"`
	Type     string    `xml:"type,attr"`
	Children []Element `xml:"complexType>sequence>element"`
}

// Function to parse the XSD file
func parseXSD(filePath string) (XSD, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return XSD{}, fmt.Errorf("failed to read file: %w", err)
	}

	//fmt.Println(string(data)) // Add this line to debug the content of the XSD file

	var xsd XSD
	err = xml.Unmarshal(data, &xsd)
	if err != nil {
		return XSD{}, fmt.Errorf("failed to unmarshal XML: %w", err)
	}
	return xsd, nil
}

// Function to generate Mustache template recursively
func generateTemplate(xsd XSD) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

	// Check if there are any elements
	if len(xsd.Elements) == 0 {
		return sb.String() // Return an empty template if no elements are found
	}

	if len(xsd.Elements[0].Children) > 0 {
		sb.WriteString("{{#" + xsd.Elements[0].Name + "}}\n")
	}

	sb.WriteString("<" + xsd.Elements[0].Name + ">\n")
	for _, element := range xsd.Elements {
		generateElementTemplate(&sb, element, "")
	}
	sb.WriteString("</" + xsd.Elements[0].Name + ">\n")

	if len(xsd.Elements[0].Children) > 0 {
		sb.WriteString("{{/" + xsd.Elements[0].Name + "}}\n")
	}

	return sb.String()
}

// Recursive function to generate template for each element
func generateElementTemplate(sb *strings.Builder, element Element, parentName string) {
	if parentName != "" {
		sb.WriteString("{{#" + parentName + "_" + element.Name + "}}\n")
		sb.WriteString("<" + element.Name + ">\n")
	}

	for _, child := range element.Children {
		if len(child.Children) > 0 { // Check if the child has its own children (complex type)
			generateElementTemplate(sb, child, element.Name) // Recursive call for nested elements
		} else {
			sb.WriteString("<" + child.Name + ">{{" + element.Name + "_" + child.Name + "}}</" + child.Name + ">\n")
		}
	}

	if parentName != "" {
		sb.WriteString("</" + element.Name + ">\n")
		sb.WriteString("{{/" + parentName + "_" + element.Name + "}}\n")
	}
}

// Define the structure for the Workato schema
type WorkatoField struct {
	Name        string         `json:"name"`
	Label       string         `json:"label,omitempty"`
	Type        string         `json:"type,omitempty"`
	Of          string         `json:"of,omitempty"`
	Optional    bool           `json:"optional,omitempty"`
	ControlType string         `json:"control_type,omitempty"`
	Properties  []WorkatoField `json:"properties,omitempty"`
}

// Function to generate Workato Schema JSON
func generateWorkatoSchema(xsd XSD) ([]WorkatoField, error) {
	var fields []WorkatoField

	for _, element := range xsd.Elements {
		workatoField := WorkatoField{
			Name:     element.Name,
			Label:    element.Name,
			Type:     mapXSDTypeToWorkatoType(element.Type), // Assuming element.Type is available
			Optional: true,                                  // Set to true or false based on your logic
		}

		// If the element has children, treat it as an object with properties
		if len(element.Children) > 0 {
			workatoField.Type = "array"
			workatoField.Of = "object"
			workatoField.Properties = generateWorkatoSchemaForChildren(element.Children, workatoField.Name)
		}

		fields = append(fields, workatoField)
	}

	return fields, nil
}

// Helper function to map XSD types to Workato types
func mapXSDTypeToWorkatoType(xsdType string) string {
	switch xsdType {
	case "xs:string":
		return "string"
	case "xs:dateTime":
		return "date_time"
	case "xs:boolean":
		return "boolean"
	case "xs:integer":
		return "integer"
	case "xs:float", "xs:double", "xs:decimal":
		return "number"
	default:
		return "string" // Default to string if type is unknown
	}
}

// Function to generate Workato Schema for child elements
func generateWorkatoSchemaForChildren(children []Element, parent string) []WorkatoField {
	var properties []WorkatoField
	var fieldName = ""
	for _, child := range children {
		if parent == "" {
			fieldName = child.Name
		} else {
			fieldName = parent + "_" + child.Name
		}
		workatoField := WorkatoField{
			Name:     fieldName,
			Label:    fieldName,
			Type:     mapXSDTypeToWorkatoType(child.Type),
			Optional: true, // Set to true or false based on your logic
		}

		// If the child has its own children, treat it as an object
		if len(child.Children) > 0 {
			workatoField.Type = "array"
			workatoField.Of = "object"
			workatoField.Properties = generateWorkatoSchemaForChildren(child.Children, child.Name)
		}

		properties = append(properties, workatoField)
	}
	return properties
}

// Function to write the Workato Schema to a JSON file
func writeWorkatoSchemaToFile(schema []WorkatoField, outputFile string) error {
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling schema to JSON: %w", err)
	}

	err = os.WriteFile(outputFile, schemaJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing schema to file: %w", err)
	}

	return nil
}

func main() {
	// Command line flag for input file
	inputFile := flag.String("i", "", "Path to the XSD file")
	flag.Parse()

	// Parse the XSD file
	xsd, err := parseXSD(*inputFile)
	if err != nil {
		fmt.Println("Error parsing XSD:", err)
		return
	}

	// Generate Mustache template
	template := generateTemplate(xsd)

	// Output file path: change the extension to .template
	templateOutputFile := strings.TrimSuffix(strings.ToLower(*inputFile), ".xsd") + ".template" // Updated

	// Write the template to a file
	err = os.WriteFile(templateOutputFile, []byte(template), 0644)
	if err != nil {
		fmt.Println("Error writing template file:", err)
		return
	}
	fmt.Println("Template generated successfully:", templateOutputFile)

	// Generate Workato Schema
	workatoSchema, err := generateWorkatoSchema(xsd)
	if err != nil {
		fmt.Println("Error generating Workato Schema:", err)
		return
	}

	// Write the Workato Schema to a file
	workatoSchemaJSONoutputFile := strings.TrimSuffix(strings.ToLower(*inputFile), ".xsd") + "-schema.json" // Updated
	err = writeWorkatoSchemaToFile(workatoSchema, workatoSchemaJSONoutputFile)
	if err != nil {
		fmt.Println("Error writing Workato Schema to file:", err)
		return
	}

	fmt.Println("Workato Schema generated successfully:", workatoSchemaJSONoutputFile)
}
