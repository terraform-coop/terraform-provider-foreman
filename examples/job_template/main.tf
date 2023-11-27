resource "foreman_jobtemplate" "testjob" {
  name = "Test job template from TF provider"
  job_category = "Testing"
  template = "echo '<%= input(\"my_input2\") %>'\n/usr/bin/<%= my_exe_name %>"
  provider_type = "script"

  template_inputs {
      name = "my_input2"
      default = "testInput123"
  }

  template_inputs {
    name = "my_input1"
    default = "abc"
  }
}