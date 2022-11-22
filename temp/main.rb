require "json"
require "securerandom"

# move recipes into separate JSON files
# recipes_file = File.read("scripts/recipes.json")
# JSON.parse(recipes_file).each do |recipe|
#   puts "processing #{recipe["name"]}"
#   filename = "pkg/classification/db/recipes/#{recipe["name"].downcase.gsub(/(\s|\/|\(|\)|\.|\,|\?)/, "_").gsub(/_{2,}/, "_").gsub(/_$/, "")}.json"
#   File.open(filename, "w") do |new_file|
#     new_file << {metadata: {version: "1.0"}}.merge(recipe).to_json
#   end
# end

# move data_category into separate JSON files
# data_categories_file = File.read("scripts/data_categories.json")
# JSON.parse(data_categories_file).each do |data_cat|
#   puts "processing #{data_cat["name"]}..."
#   filename = "pkg/classification/db/data_categories/#{data_cat["name"].downcase.gsub(" ", "_")}.json"
#   File.open(filename, "w") do |new_file|
#     new_file << {
#       metadata: {
#         version: "1.0"
#       }
#     }.merge(data_cat).to_json
#   end
# end

# YAML to JSON conversion:
#   ruby -ryaml -rjson -e "puts YAML.load_file('path-to-your-file.yml').to_json"

data_categories_file = File.read("scripts/data_categories.json")
data_categories_hash = JSON.parse(data_categories_file)
# move data types into separate JSON files
data_types_file = File.read("scripts/data_types.json")

def find_category_uuid(name, data_categories_hash)
  res = ""

  data_categories_hash.each do |data_category|
    if data_category["name"] == name
      res = data_category["uuid"]
      break
    end
  end

  res
end

JSON.parse(data_types_file).each do |pattern|
  puts "processing #{pattern["data_category_name"]}"
  filename = "pkg/classification/db/data_types/#{pattern["data_category_name"].downcase.gsub(" ", "_")}.json"
  File.open(filename, "w") do |new_file|
    new_file << {
      metadata: {version: "1.0"},
      name: pattern["data_category_name"],
      uuid: pattern["uuid"],
      category_uuid: find_category_uuid(pattern["default_category"], data_categories_hash),

    }.to_json
  end
end

# move data type classification patterns to separate JSON files
# data_type_classification_patterns_file = File.read("scripts/data_type_classification_patterns.json")
# JSON.parse(data_type_classification_patterns_file).each do |pattern|
#   puts "processing #{pattern["friendly_name"]}"

#   # remove spaces in object type
#   pattern["object_type"] = pattern["object_type"].map { |object_type| object_type.downcase.gsub(" ", "_") }

#   filename = "pkg/classification/db/data_type_classification_patterns/#{pattern["id"]}_#{pattern["friendly_name"].downcase.gsub(/\s+/, "_").gsub("/", "_")}.json"
#   File.open(filename, "w") do |new_file|
#     new_file << {metadata: {version: "1.0"}}.merge(pattern).to_json
#   end
# end

# move known object patterns into separate JSON files
# known_person_object_patterns_file = File.read("scripts/known_person_object_patterns.json")
# JSON.parse(known_person_object_patterns_file).each do |pattern|
#   puts "processing #{pattern["category"]}"

#   filename = "pkg/classification/db/known_person_object_patterns/#{pattern["category"].downcase.gsub(" ", "_")}.json"
#   File.open(filename, "w") do |new_file|
#     new_file << {metadata: {version: "1.0"}}.merge(pattern).to_json
#   end
# end