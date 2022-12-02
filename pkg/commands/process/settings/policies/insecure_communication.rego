package bearer.insecure_communication

import future.keywords

sensitive_data_group_uuid := "f6a0c071-5908-4420-bac2-bba28d41223e"

medium[item] {
    some data_type in input.dataflow.data_types
    some data_category in input.data_categories
    data_category.uuid == data_type.category_uuid
    data_category.group_uuid == sensitive_data_group_uuid

    some detector in input.dataflow.risks
    detector.detector_id == input.policy_id

    location = detector.locations[_]
    item := {
      "category_group": data_category.group_name,
      "filename": location.filename,
      "line_number": location.line_number,
      "parent_line_number": location.parent.line_number,
      "parent_content": location.parent.content
    }
}
