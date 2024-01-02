require_relative 'tomlparser-utils'
require "colorize"

RSpec.describe 'When passing a string to the splitMetricsRegex method' do
  it 'correctly splits metrics in valid segments' do
    input_string = "(go_one|process_cpu_two|process_max_three)|process_virtual_four|process_(cpu|max|resident|virtual|open)_.*|apiserver_request_total"

    expected_result = [
      "(go_one|process_cpu_two|process_max_three)",
      "process_virtual_four",
      "process_(cpu|max|resident|virtual|open)_.*",
      "apiserver_request_total"
    ]

    expect(splitMetricsRegex(input_string)).to eq(expected_result)
  end

  it 'correctly splits metrics in valid segments with multiple parentheses' do
    input_string = "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)|kube_deployment_(status_(condition|replicas(_(available|updated|ready)))|labels|spec_replicas)|kube_cronjob_(status_(last_schedule_time))|kube_job_status_(failed|start_time)*"
    expected_result = [
      "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)",
      "kube_deployment_(status_(condition|replicas(_(available|updated|ready)))|labels|spec_replicas)",
      "kube_cronjob_(status_(last_schedule_time))",
      "kube_job_status_(failed|start_time)*"
    ]

    expect(splitMetricsRegex(input_string)).to eq(expected_result)
  end
end

$exclusionsRegex=/\A(\()?(?:go|process(?!_start_time_seconds)|(.*\|)*(process|go)(_)?\|)/
RSpec.describe 'When passing a string to the excludeMetricsRegex method' do
  let(:exclusions_regex) { $exclusionsRegex }
  it 'correctly excludes metrics from simple keep list' do
    input_string = "go_one|process_cpu_two|apiserver_request_total"

    expected_result = "apiserver_request_total"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly excludes metrics from keep list with multiple parentheses and starting with (go_' do
    input_string = "(go_one|process_cpu_two|pod_security_exemptions_total)|process_virtual_four|process_(cpu|max|resident|virtual|open)_.*|apiserver_request_total"

    expected_result = "apiserver_request_total"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly excludes metrics from keep list with multiple parentheses and metrics to exclude in the middle of the string' do
    input_string = "(etcd_metric|process|pod_metric)_cpu_two|process_virtual_four|process_(cpu|max|resident|virtual|open)_.*|apiserver_request_total"

    expected_result = "apiserver_request_total"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly excludes metrics from keep list with multiple parentheses and metrics to exclude in the middle of the string including underscores' do
    input_string = "(etcd_metric_|process_|go_|pod_metric_)cpu_two|process_virtual_four|process_(cpu|max|resident|virtual|open)_.*|apiserver_request_total|rest_client_exec_plugin_ttl_seconds"

    expected_result = "apiserver_request_total|rest_client_exec_plugin_ttl_seconds"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly will not exclude metrics from keep list if keywords are in the middle of a metric name' do
    input_string = "test_process_metric|metric_go_keep"

    expected_result = "test_process_metric|metric_go_keep"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly will not exclude metrics from keep list if keywords are not in the input string' do
    input_string = "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)|kube_deployment_(status_(condition|replicas(_(available|updated|ready)))|labels|spec_replicas)|kube_cronjob_(status_(last_schedule_time))|kube_job_status_(failed|start_time)*"
    expected_result = "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)|kube_deployment_(status_(condition|replicas(_(available|updated|ready)))|labels|spec_replicas)|kube_cronjob_(status_(last_schedule_time))|kube_job_status_(failed|start_time)*"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end

  it 'correctly will not exclude process_start_time_seconds from keep list' do
    input_string = "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)|process_start_time_seconds|process_max_fds|go_cpu_classes_gc_mark_(assist_cpu_seconds_total|dedicated_cpu_seconds_total)"
    expected_result = "kube_pod_container_status_(restarts_total|waiting_reason|last_terminated_reason)|process_start_time_seconds"

    expect(excludeMetricsRegex(input_string, exclusions_regex)).to eq(expected_result)
  end
end