# XSD to Workato Schema and XML Message Template Generator

## Overview

This Go application is designed to parse XML Schema Definition (XSD) files and generate corresponding Mustache templates and Workato schemas in JSON format. The application reads an XSD file, extracts its elements and their properties, and outputs a structured template and schema that can be used for further processing or integration.

## Features

- **Parse XSD Files**: Reads and parses XSD files to extract elements and their attributes.
- **Generate Mustache Templates**: Creates Mustache templates based on the parsed XSD elements, allowing for dynamic content generation.
- **Generate Workato Schemas**: Converts the parsed XSD elements into Workato-compatible JSON schemas, which can be used in automation workflows.
- **Validation**: Basic validation of the generated Mustache templates to ensure they contain the necessary structure.

## Usage

To run the application, use the following command:
