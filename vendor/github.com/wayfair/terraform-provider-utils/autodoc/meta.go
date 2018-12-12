package autodoc

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// NOTE(ALL): If you make modifications to the metadata, be
//   sure to update the documentation! This includes:
//
//   * The autodoc tool documentation in docs/autodoc.md

// Metadata tags. These can be set in the description field of the schema
// to provide extra information when generating docs. Tags should be
// separated by a space (ie: "@FOO @BAR" not "@FOO@BAR").  If a tag supports
// a value, the value for that tag should come after the tag, separated by
// a space (ie: "@FOO value for foo tag @BAR")
const (
	// The meta attribute. If an attribute has this name, autodoc will interpret
	// it as a special tag and parses its description for more resource-level
	// metadata information. This attribute is never included in the docs.
	MetaAttribute = "__meta__"
	// Metadata tag to signal that this resource cannot be created. This should
	// be in the description of the meta attribute. This tag does
	// not accept a value. The default behavior assumes the resource can be
	// created. This over-rides that behavior.
	MetaNotCreatable = "@NOTCREATABLE"
	// Metadata tag to signal that this resource cannot be deleted. This should
	// be in the description of the meta attribute. This tag does not
	// accept a value. The default behavior assumes the resource can be
	// deleted. This over-rides that behavior.
	MetaNotDeletable = "@NOTDELETABLE"
	// Metadata tag to signal that this resource cannot be updated. This should
	// be in the description of the meta attribute. This tag does
	// not accept a value. The default behavior assumes the resource can be
	// updated. this over-rides that behavior.
	MetaImmutable = "@IMMUTABLE"
	// Metadata tag that provides more information. This should be in the
	// description of the meta attribute. This tag accepts a value corresponding
	// the summary for this resource.
	MetaSummary = "@SUMMARY"
	// Metadata tag to provide an example value for the resource's argument.
	// This should be in the description for one of the resource's arguments.
	// This tag accepts a value corresponding to an example value for this
	// argument.
	MetaExample = "@EXAMPLE"
	// Metadata tag to denote a resource's attributed as unexported.  Unexported
	// attributes are not exposed to other resources. The default behavior is
	// to assume the attribute is exported. This will over-ride that behavior.
	// This tag does not accept a value.
	MetaUnexported = "@UNEXPORTED"
)

// -----------------------------------------------------------------------------
// Metadata Definition
// -----------------------------------------------------------------------------

// NOTE(ALL): If you make modifications to the metadata, be
//   sure to update the documentation! This includes:
//
//   * The autodoc tool documentation in docs/autodoc.md

// Metadata information for the resource
type meta struct {
	// Whether or not this resource supports create
	Uncreatable bool
	// Whether or not this resource supports delete
	Undeletable bool
	// Whether or not this resource supports update
	Immutable bool
	// Summary of the resource
	Summary string
}

// -----------------------------------------------------------------------------
// Metadata Utility Functions
// -----------------------------------------------------------------------------

// parseMeta parses the metadata from the resource map. The metadata is defined
// in the metadata attribute's description.
func parseMeta(schemaMap map[string]*schema.Schema) meta {
	meta := meta{}
	for attrName, attrSchema := range schemaMap {
		if attrName == MetaAttribute {
			meta.Uncreatable = strings.Contains(attrSchema.Description, MetaNotCreatable)
			meta.Undeletable = strings.Contains(attrSchema.Description, MetaNotDeletable)
			meta.Immutable = strings.Contains(attrSchema.Description, MetaImmutable)
			meta.Summary = parseMetaValue(attrSchema.Description, MetaSummary)
			break
		}
	}
	return meta
}

// parseMetaValue parses a schema description string for metadata tag
// and returns the value associated with that tag.
func parseMetaValue(descr string, metaTag string) string {
	value := ""

	// find where the meta tag exists in the description string. If the meta
	// tag cannot be found, return
	metaTagIdx := strings.Index(descr, metaTag)
	if metaTagIdx < 0 {
		return value
	}

	// take a substring of the description starting with the end of the
	// meta tag. This is our known metadata value
	value = descr[metaTagIdx+len(metaTag):]
	// value is the empty string, return
	valueLen := len(value)
	if valueLen == 0 {
		return value
	}

	// Retrieve the ending location of the value. This is either the index of
	// next meta tag, or the end of the string
	var valueEndIdx int
	if valueEndIdx = nextMetaTagIndex(value); valueEndIdx < 0 {
		valueEndIdx = valueLen - 1
	}

	// The value text will be the substring to this index. Trim any
	// leading and trailing whitespace.
	//
	// If the value goes to the end of the string, add 1 to the end index when
	// taking the substring. The index operator [:idx] is exclusive on the right
	// bound. We don't want to miss this character at the end of the string.  In
	// the cases where we don't hit the end of the string, we can just go to
	// valueEndIdx. If we were to add 1 in this case, we'd include the '@' of
	// the next tag.
	if valueEndIdx == valueLen-1 {
		return strings.TrimSpace(value[:valueEndIdx+1])
	}
	return strings.TrimSpace(value[:valueEndIdx])
}

// nextMetaTagIndex returns the character index of the first encountered
// meta tag. -1 is returned if no valid meta tags are found.
func nextMetaTagIndex(value string) int {
	// Assume we are not going to encounter another meta tag and will run
	// to the end of the string
	valueLen := len(value) - 1
	valueEndIdx := valueLen
	// Move the index back if we encounter another metadata tag. We want the
	// text immediately following the summary tag up to the next metadata tag
	metaTags := []string{
		MetaNotCreatable,
		MetaNotDeletable,
		MetaImmutable,
		MetaSummary,
		MetaExample,
		MetaUnexported,
	}
	for _, tag := range metaTags {
		if endIdx := strings.Index(value, tag); endIdx != -1 && endIdx < valueEndIdx {
			valueEndIdx = endIdx
		}
	}
	// No other meta tags found
	if valueEndIdx == valueLen {
		return -1
	}
	// Hit another tag, give its position
	return valueEndIdx
}

// stripMeta removes any metadata tags from a schema description and their
// associated values.
func stripMeta(descr string) string {
	descrCleaned := descr

	// for meta tags that do not accept a value, just remove them
	metaTagsNoValue := []string{
		MetaNotCreatable,
		MetaNotDeletable,
		MetaImmutable,
		MetaUnexported,
	}
	for _, tag := range metaTagsNoValue {
		descrCleaned = strings.Replace(descrCleaned, tag, "", 1)
	}

	// for meta tags that accept a value, remove the tag and its value
	metaTagsValue := []string{
		MetaSummary,
		MetaExample,
	}
	for _, tag := range metaTagsValue {
		tagLen := len(tag)
		metaBeginIdx := strings.Index(descrCleaned, tag)
		// tag not found - skip
		if metaBeginIdx < 0 {
			continue
		}
		// starting at the end of tag, search for start of another tag
		metaEndIdx := nextMetaTagIndex(descrCleaned[metaBeginIdx+tagLen:])
		// Meta value goes to the end of the string. Create a substring
		// to the beginning of the meta tag - effectively cutting the meta
		// tag and value out
		if metaEndIdx < 0 {
			descrCleaned = descrCleaned[:metaBeginIdx]
			continue
		}
		// offset the end index - the next meta tag index was calculated from the
		// value of the tag
		metaEndIdx += metaBeginIdx + tagLen
		// Meta value is somewhere inside the string. Concatenate the string
		// to the beginning of the meta tag with the string starting with the
		// end of the meta tag value - effectively cutting the meta tag and
		// value out.
		descrCleaned = descrCleaned[:metaBeginIdx] + descrCleaned[metaEndIdx:]
	}

	return strings.TrimSpace(descrCleaned)
}
