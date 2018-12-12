package autodoc

import (
	"os"
	"path/filepath"
	"text/template"
)

// NOTE(ALL): If you make modifications to the template associations, be
//   sure to update the documentation! This includes:
//
//   * The package comment in autodoc.go
//   * The Usage() function in autodoc.go
//   * The autodoc tool documentation in docs/autodoc.md

// Template associations for different output files. The template extension
// from the command-line arguments will be appended to these files when
// selecting the correct template to use.
const (
	// Template file for mkdocs.yml
	mkdocsYmlTemplate = "mkdocs.yml"
	// Template file for godoc.md
	godocMdTemplate = "godoc.md"
	// Template file for all provider resources
	resourceMdTemplate = "resource.md"
	// Template file for all provider data sources
	dataSourceMdTemplate = "datasource.md"
	// Template file for the provider itself
	providerMdTemplate = "index.md"
)

// The type of schema that is being documented
const (
	// Provider schema map
	typeProvider = iota
	// Resource schema map
	typeResource
	// Data source schema map
	typeDataSource
)

// -----------------------------------------------------------------------------
// Template Data Structs
// -----------------------------------------------------------------------------

// Template data needed to generate mkdocs.yml
type mkdocsYmlData struct {
	// The docs_dir - location where documentation files are generated to
	DocsDir string
	// List of provider resources
	Resources []string
	// List of provider data sources
	DataSources []string
}

// Template data needed to generate a provider, resource, or data source
// documentation file.
type schemaDocData struct {
	// Constants map. Go does not expose constants to runtime processes (like
	// templates, reflect) they only exist in the compiled binary. For this
	// reason, in order to expose them to our template we will pass them as a
	// map. The string is the name of the constant, the value is the value of
	// that constant.
	Constants map[string]interface{}
	// The type of schema. This denotes whether this is a provider, resource,
	// or data source schema. This should be one of the typeXxx constants.
	SchemaType int
	// Name of the resource
	Name string
	// Metadata information about the resource
	Meta meta
	// List of resource's exported schema attributes
	Attributes []schemaAttribute
	// List of resource's schema arguments
	Arguments []schemaArgument
}

// Template data representing an attribute of a resource
type schemaAttribute struct {
	// Name of the attribute
	Name string
	// Type of the attribute, in string form
	Type string
	// Description of the attribute
	Description string
}

// Template data representing an argument of a resource
type schemaArgument struct {
	// Name of the argument
	Name string
	// Type of the argument, in string form
	Type string
	// An example value for this argument
	Example string
	// Whether or not this argument is optional
	Optional bool
	// Whether or not a modification to this argument causes the resource
	// to be destroyed and then recreated.
	ForceNew bool
	// Description of the argument
	Description string
	// A list of strings representing the conflicting schema arguments. If
	// a schema has ConflictsWith set, this means only one of that argument
	// or the list of arguments in the ConflictsWith definition can be set
	// in the config.
	ConflictsWith []string
}

// -----------------------------------------------------------------------------
// Template Utility Functions
// -----------------------------------------------------------------------------

// parseTemplates recursively searches the templates directory (from
// parsedArgs.templatesDir) for template files (from
// parsedArgs.templateFileExt). Returns the text template reference on
// success or an error if one was encountered.
func parseTemplates(args parsedArgs) (*template.Template, error) {
	t := template.New("")

	// walk the templates directory, if we encounter any sub directories we load
	// the template files in them and keep walking down
	var parseErr error
	walkErr := filepath.Walk(args.templatesDir, func(path string, info os.FileInfo, err error) error {
		pathGlob := filepath.Join(path, "*"+args.templateFileExt)
		if info.IsDir() {
			_, parseErr = t.ParseGlob(pathGlob)
			return parseErr
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return t, nil
}
