
# foreman_jobtemplate


Foreman representation of a job template.


## Example Usage

```
# Autogenerated example with required keys
data "foreman_jobtemplate" "example" {
  name = "change content sources"
}
```


## Argument Reference

The following arguments are supported:

- `name` - (Required) job template name.


## Attributes Reference

The following attributes are exported:

- `description` - 
- `description_format` - 
- `job_category` - 
- `locked` - 
- `name` - job template name.
- `provider_type` - 
- `snippet` - 
- `template` - The template content itself
- `template_inputs` - 

