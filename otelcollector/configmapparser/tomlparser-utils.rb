#!/usr/local/bin/ruby
# frozen_string_literal: true

require_relative "ConfigParseErrorLogger"

if (!ENV["OS_TYPE"].nil? && ENV["OS_TYPE"].downcase == "linux")
  require "re2"
end

# RE2 is not supported for windows
def isValidRegex_linux(str)
  begin
    # invalid regex example -> 'sel/\\'
    re2Regex = RE2::Regexp.new(str)
    return re2Regex.ok?
  rescue => errorStr
    ConfigParseErrorLogger.logError(LOGGING_PREFIX, "Exception while validating regex for target metric keep list - #{errorStr}, regular expression str - #{str}")
    return false
  end
end

def isValidRegex_windows(str)
  begin
    # invalid regex example -> 'sel/\\'
    re2Regex = Regexp.new(str)
    return true
  rescue => errorStr
    ConfigParseErrorLogger.logError(LOGGING_PREFIX, "Exception while validating regex for target metric keep list - #{errorStr}, regular expression str - #{str}")
    return false
  end
end

def isValidRegex(str)
  if ENV["OS_TYPE"] == "linux"
    return isValidRegex_linux(str)
  else
    return isValidRegex_windows(str)
  end
end

# Some targets have metrics that won't be exposed so we need to make sure to remove them from the keep list
def excludeMetricsRegex(keepListRegex,excludeListRegex)
  adjustedKeepList = ""
  begin
    metrics = splitMetricsRegex(keepListRegex)
    filtered_metrics = metrics.select { |metric| !metric.match(excludeListRegex) }
    #filtered_metrics = metrics.reject { |metric| metric.match(/\A(\()?(?:go|process|(.*\|)*(process|go)(_)?\|)/) }
    if !filtered_metrics.nil? && !filtered_metrics.empty?
      adjustedKeepList = filtered_metrics.join('|')
    end
  end
  return adjustedKeepList
end

def splitMetricsRegex(metrics_regex)
  if metrics_regex.start_with?("(") && metrics_regex.end_with?(")")
    metrics_regex = metrics_regex[1..-2]
  end

  segments = []
  current_segment = ""
  parenthesis_depth = 0

  metrics_regex.chars.each do |char|
    if char == '|'
      # Don't split '|' inside parenthesis
      if parenthesis_depth == 0
        segments << current_segment
        current_segment = ""
      else
        current_segment += char
      end
    else
      current_segment += char
      if char == '('
        parenthesis_depth += 1
      elsif char == ')'
        parenthesis_depth -= 1
      end
    end
  end

  # Add the last segment to the result
  segments << current_segment

  return segments
end