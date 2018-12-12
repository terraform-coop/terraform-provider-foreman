package autodoc

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform/helper/schema"
)

// -----------------------------------------------------------------------------
// Goroutine Data Structs - These structures are passed to each goroutine.  The
//   data is parsed into one of the template data structures before executing
//   the template.
// -----------------------------------------------------------------------------

// Base goroutine input data structure.  All of the documentation generating
// goroutines will have access to this information to operate properly.
type goroutineBase struct {
	// Path to the output file
	outFile string
	// Reference to the loaded & parsed text templates tree
	template *template.Template
	// Name of the template to use to generate the output file
	templateName string
	// Bidirectional error channel for communication to main goroutine. This
	// should return nil if no errors are encountered when generating the
	// documentation.  Otherwise, the main goroutine will receive an error
	// from this channel.
	errChan chan error
}

// Represents the mkdocs.yml document. This information is passed to the
// goroutine generating mkdocs.yml
type mkdocsYmlDoc struct {
	// Contains base goroutine information
	goroutineBase
	// Includes a reference to the Terraform provider
	provider *schema.Provider
	// Includes a reference to the command line arguments
	args parsedArgs
}

// Represents a markdown schema document. This information is passed to the
// goroutine generating the provider, resource, and data source documentation.
type schemaDoc struct {
	// Contains base goroutine information
	goroutineBase
	// The type of schema. This denotes whether this is a provider, resource,
	// or data source schema. This should be one of the typeXxx constants.
	schemaType int
	// Name of this resource object
	name string
	// Include a reference to the schema to be documented
	schema map[string]*schema.Schema
}

// -----------------------------------------------------------------------------
// Documentation generator functions - these are run by goroutines spawned
// by autodoc.Document()
// -----------------------------------------------------------------------------

// generateSchemaDoc generates documentation for a resource's schema map (ie:
// the reosurce's 'Schema' attribute). This can be a schema map for a provider,
// resource, or a data.
func generateSchemaDoc(d schemaDoc) {
	// template data
	data := schemaDocData{
		Constants: map[string]interface{}{
			"TypeProvider":   typeProvider,
			"TypeResource":   typeResource,
			"TypeDataSource": typeDataSource,
		},
		SchemaType: d.schemaType,
		Name:       d.name,
		Meta:       parseMeta(d.schema),
		Attributes: schemaAttributes(d.schema),
		Arguments:  schemaArguments(d.schema),
	}
	// sort argument and attributes list alphabetically for easier reading
	sort.Slice(data.Arguments, func(i, j int) bool {
		return data.Arguments[i].Name < data.Arguments[j].Name
	})
	sort.Slice(data.Attributes, func(i, j int) bool {
		return data.Attributes[i].Name < data.Attributes[j].Name
	})

	// requested template should exist and be defined
	if d.template.Lookup(d.templateName) == nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [%s]. Template [%s] "+
				"does not exist or is not defined.",
			d.outFile,
			d.templateName,
		)
		return
	}

	// open output file
	fd, fileErr := openFile(d.goroutineBase)
	defer fd.Close()
	if fileErr != nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [%s]. Failed to get file descriptor. "+
				"Error: [%s]",
			d.goroutineBase.outFile,
			fileErr.Error(),
		)
		return
	}

	// Execute template with supplied data, dump output to our file descriptor
	templateErr := d.template.ExecuteTemplate(
		fd,
		d.templateName,
		data,
	)

	// Signal error back to main goroutine
	d.errChan <- templateErr
}

// generateGodocMd generates the wrapper documentation file that serves as a
// viewport to the godoc.
func generateGodocMd(d goroutineBase) {
	// requested template should exist and be defined
	if d.template.Lookup(d.templateName) == nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [%s]. Template [%s] "+
				"does not exist or is not defined.",
			d.outFile,
			d.templateName,
		)
		return
	}

	// open output file
	fd, fileErr := openFile(d)
	defer fd.Close()
	if fileErr != nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [%s]. Failed to get file descriptor. "+
				"Error: [%s]",
			d.outFile,
			fileErr.Error(),
		)
		return
	}

	// Execute template with supplied data, dump output to our file descriptor
	templateErr := d.template.ExecuteTemplate(
		fd,
		d.templateName,
		nil,
	)

	// Signal error back to main goroutine
	d.errChan <- templateErr
}

// generateMkdocsYml genreates the mkdocs.yml file which configures the
// mkdocs build.
func generateMkdocsYml(d mkdocsYmlDoc) {
	// template data
	data := mkdocsYmlData{
		DocsDir: d.args.docsDir,
	}

	// requested template should exist and be defined
	if d.template.Lookup(d.templateName) == nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [%s]. Template [%s] "+
				"does not exist or is not defined.",
			d.outFile,
			d.templateName,
		)
		return
	}

	// provider reference should not be nil
	if d.provider == nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [mkdocs.yml]. Provider reference is nil.",
		)
		return
	}

	// get the list of resources, data sources from the provider schema
	for resourceName, _ := range d.provider.ResourcesMap {
		data.Resources = append(data.Resources, resourceName)
	}
	for dataSourceName, _ := range d.provider.DataSourcesMap {
		data.DataSources = append(data.DataSources, dataSourceName)
	}

	// sort the resource, data source list by name for easier reading
	sort.Slice(data.Resources, func(i, j int) bool {
		return data.Resources[i] < data.Resources[j]
	})
	sort.Slice(data.DataSources, func(i, j int) bool {
		return data.DataSources[i] < data.DataSources[j]
	})

	// open output file
	fd, fileErr := openFile(d.goroutineBase)
	defer fd.Close()
	if fileErr != nil {
		d.errChan <- fmt.Errorf(
			"Cannot generate [mkdocs.yml]. Could not get file descriptor. "+
				"Error: [%s]",
			fileErr.Error(),
		)
		return
	}

	// Execute template with supplied data, dump output to our file descriptor
	templateErr := d.template.ExecuteTemplate(
		fd,
		d.templateName,
		data,
	)

	// Signal error back to main goroutine
	d.errChan <- templateErr
}

// -----------------------------------------------------------------------------
// Documentation Utility Functions
// -----------------------------------------------------------------------------

// schemaAttributes scans all the schema attributes and parses them into
// a list of exported schema attributes
func schemaAttributes(schemaMap map[string]*schema.Schema) []schemaAttribute {
	attrs := []schemaAttribute{}
	for attrName, attrSchema := range schemaMap {
		// skip the meta attribute
		if attrName == MetaAttribute {
			continue
		}
		// if the attribute is tagged as unexported, do not include it in the
		// attribute list
		if strings.Contains(attrSchema.Description, MetaUnexported) {
			continue
		}
		attr := schemaAttribute{
			Name:        attrName,
			Type:        schemaType(attrSchema),
			Description: stripMeta(attrSchema.Description),
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// schemaArguments scans all the schema attributes and parses them into
// a list of schema arguments
func schemaArguments(schemaMap map[string]*schema.Schema) []schemaArgument {
	args := []schemaArgument{}
	for argName, argSchema := range schemaMap {
		// skip the meta attribute
		if argName == MetaAttribute {
			continue
		}
		// Only consider required or optional attributes. Computed attributes
		// aren't given values in the HCL and therefore are not considered
		// an agrument
		//
		// It is valid in Terraform for an attribute to be Computed and
		// Optional.  In this case, if it supplied a value, it needs to be
		// treated as an argument. However, it is not possible for an argument
		// to be both Computed and Required. Terraform does not allow this
		// behavior.
		if argSchema.Computed && !argSchema.Optional {
			continue
		}
		arg := schemaArgument{
			Name:          argName,
			Type:          schemaType(argSchema),
			Example:       parseMetaValue(argSchema.Description, MetaExample),
			Description:   stripMeta(argSchema.Description),
			Optional:      argSchema.Optional,
			ForceNew:      argSchema.ForceNew,
			ConflictsWith: argSchema.ConflictsWith,
		}
		args = append(args, arg)
	}
	return args
}

// schemaType parses the schema definition for its type and returns a string
// representation of the type with markdown formatting. If the type is simple
// (ie: schema.TypeBool), then the output string will just be that escaped
// type. For complex types (ie: schema.TypeList) it will return the type
// and the element type. Unrecognized types will return 'unknown'.
func schemaType(s *schema.Schema) string {
	switch s.Type {
	case schema.TypeBool:
		return "`schema.TypeBool`"
	case schema.TypeInt:
		return "`schema.TypeInt`"
	case schema.TypeFloat:
		return "`schema.TypeFloat`"
	case schema.TypeString:
		return "`schema.TypeString`"
	case schema.TypeList:
		if s.Elem == nil {
			return "`schema.TypeList`"
		}
		if _, ok := s.Elem.(*schema.Resource); ok {
			return "`schema.TypeList` of `schema.Resource`"
		}
		if elem, ok := s.Elem.(*schema.Schema); ok {
			return fmt.Sprintf(
				"`schema.TypeList` of %s",
				schemaType(elem),
			)
		}
		return "`schema.TypeList` of `unknown`"
	case schema.TypeSet:
		if s.Elem == nil {
			return "`schema.TypeSet`"
		}
		if _, ok := s.Elem.(*schema.Resource); ok {
			return "`schema.TypeSet` of `schema.Resource`"
		}
		if elem, ok := s.Elem.(*schema.Schema); ok {
			return fmt.Sprintf(
				"`schema.TypeSet` of %s",
				schemaType(elem),
			)
		}
		return "`schema.TypeList` of `unknown`"
	case schema.TypeMap:
		if s.Elem == nil {
			return "`schema.TypeMap`"
		}
		if _, ok := s.Elem.(*schema.Resource); ok {
			return "`schema.TypeMap` of `schema.Resource`"
		}
		if elem, ok := s.Elem.(*schema.Schema); ok {
			return fmt.Sprintf(
				"`schema.TypeMap` of %s",
				schemaType(elem),
			)
		}
		return "`schema.TypeMap` of `unknown`"
	default:
		return "`unknown`"
	}
}

// openFile reads the outFile of the supplied goroutineBase and attempts
// to open it for writing. If the file does not exist, it will be created.
// If the file already exists, it will be truncated when opened. An error is
// returned if the file could not be opened.
func openFile(r goroutineBase) (*os.File, error) {
	// outFile should be defined
	if r.outFile == "" {
		return nil, fmt.Errorf(
			"Cannot generate file. No outfile specified.",
		)
	}
	// attempt to open the output file for writing to dump our template. If the
	// file already exists, overwrite its contents.
	return os.OpenFile(
		r.outFile,
		// Write only, create file if doesn't exist, truncate file when opened
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0775,
	)
}
