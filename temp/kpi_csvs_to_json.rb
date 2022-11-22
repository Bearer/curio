# Script for our Rails app - for schema KPI testing

# Reads in the input and output CSVs for a language,
# deduplicates their content, and constructs a JSON file for ingestion by
# our Curio schema testhelper

require "csv"
require "json"

%w(
  ruby
  java
  csharp
  go
  php
  python
  typescript
  javascript
).each do |lang|
  input_file = "#{lang}-input.csv"
  output_file = "#{lang}-output.csv"

  CSV.open("#{lang}-input-deduped.csv", "w") do |csv|
    CSV.read(input_file).uniq do |row|
      "#{row[1]}-#{row[8]}-#{row[5]}"
    end.each do |row|
      csv << row
    end
  end

  CSV.open("#{lang}-output-deduped.csv", "w") do |csv|
    CSV.read(output_file).uniq do |row|
      "#{row[0]}-#{row[1]}"
    end.each do |row|
      csv << row
    end
  end

  results = {}

  deduped_input_file = "#{lang}-input-deduped.csv"
  deduped_output_file = "#{lang}-output-deduped.csv"

  def translate_reason(reason)
    case reason
    when "invalid_object_but_db_identifier"
      "known_database_identifier"
    when "valid_object_and_known_pattern"
      "known_pattern"
    else
      reason
    end
  end

  output_results = {}
  File.open(deduped_output_file) do |file|
    CSV.new(file, headers: true).each do |row|
      valid_fields = false

      output_results[row.fetch("object_name")]
      field_name = row.fetch("field_name")
      obj = output_results.fetch(row.fetch("object_name"), {})

      true_positive = row.fetch("true_or_false") || false
      if field_name == nil || field_name == ""
        # we are dealing with an object
        obj.merge!({
          state: "valid",
          reason: translate_reason(row.fetch("reason")),
          false_positive: !true_positive
        })
      else
        obj[:fields] ||= {}
        obj[:fields][field_name] ||= {}

        valid_fields = true

        obj[:fields][field_name].merge!({
          state: "valid",
          reason: translate_reason(row.fetch("reason")),
          false_positive: !true_positive
        })
      end

      if valid_fields
        # TODO: generate reason for these fields?
        # TODO: calculate false_positive for these fields?
        obj.merge!({state: "valid"})
      end

      output_results[row.fetch("object_name")] = obj
    end
  end

  File.open(deduped_input_file) do |file|
    CSV.new(file, headers: true).each do |row|
      key = "#{row.fetch("Object name")}"

      output_result = output_results.fetch(key, {})
      if row.fetch("Field name") == nil || row.fetch("Field name") == ""
        # dealing with an object
        results[key] = results.fetch(key, {}).merge!({
          name: row.fetch("Object name"),
          type: row.fetch("Simple field type"),
          filename: row.fetch("File Name"),
          detector_type: row.fetch("Language"),
          state: output_result[:state] || "invalid",
          reason: output_result[:reason],
          false_positive: output_result[:false_positive],
        }.compact)
      else
        if results[key] == nil
          results[key] = {
            name: row.fetch("Object name"),
            filename: row.fetch("File Name"),
            detector_type: row.fetch("Language"),
            properties: {},
            state: output_result[:state],
            reason: output_result[:reason],
            false_positive: output_result[:false_positive],
          }.compact
        end

        output_field_result = output_result.fetch(:fields, {}).fetch(row.fetch("Field name"), {})
        field_key = "#{row.fetch("Field name")}"
        results[key][:properties][field_key] = {
          name: row.fetch("Field name"),
          type: row.fetch("Simple field type"),
          state: output_field_result[:state],
          reason: output_field_result[:reason],
          false_positive: output_field_result[:false_positive]
        }.compact
      end
    end
  end

  File.open("#{lang}.json", "w") do |f|
    results_arr = []
    results.each do |k, v|
      v[:properties] = v.fetch(:properties, {}).values
      results_arr << v
    end

    f.write(results_arr.to_json)
  end
end