output "env_check" {
  value = {
    # Check if TF_LOG is set (with strict=true by default)
    tf_log_exists = provider::helpers::os_check_env("TF_LOG")

    # Check if TF_LOG is set with strict=false (allows empty values)
    tf_log_exists_non_strict = provider::helpers::os_check_env("TF_LOG", false)

    # Check if a non-existent variable exists
    non_existent_var = provider::helpers::os_check_env("NON_EXISTENT_VAR")
  }
}
