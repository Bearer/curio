package bearer.cr_001

import data.bearer.common

import future.keywords

policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == input.policy_id

    data_type = detector.data_types[_]
    data_type.name != "Unique Identifier"

    location = data_type.locations[_]
    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": data.bearer.common.severity_of_datatype(data_type),
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}
