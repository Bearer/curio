package bearer.common

import future.keywords

sensitive_data_group_uuid := "f6a0c071-5908-4420-bac2-bba28d41223e"
personal_data_group_uuid := "e1d3135b-3c0f-4b55-abce-19f27a26cbb3"

severity_of_datatype(data_type) := "critical" if {
    some category in input.data_categories
    category.uuid == data_type.category_uuid

    some group in category.groups
    group.uuid == sensitive_data_group_uuid
}

severity_of_datatype(data_type) := "high" if {
    some category in input.data_categories
    category.uuid == data_type.category_uuid

    some group in category.groups
    group.uuid == personal_data_group_uuid

    every group_1 in category.groups {
        group_1.uuid != sensitive_data_group_uuid
    }
}

groups_for_datatype(data_type) := x if {
    some category in input.data_categories
    category.uuid == data_type.category_uuid

    x := {name | name := category.groups[_].name}
}

groups_for_datatypes(data_types) := groups if {
    groups := {name | name := groups_for_datatype(data_types[_])[_]}
}
