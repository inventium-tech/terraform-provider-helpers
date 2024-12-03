output "tf_log" {
  value = {
    test_value_as_is         = provider::helpers::os_get_env("TF_LOG")
    test_value_with_fallback = provider::helpers::os_get_env("TF_ENV", "test")
  }
}
