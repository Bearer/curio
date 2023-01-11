package bearer.cr_023

import data.bearer.common

import future.keywords

password_uuid := "02bb0d3a-2c8c-4842-be1c-c057f0dccd63"

rc4_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-023-5"

  location = detector.locations[_]

  item := location.filename
}

blowfish_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-023-2"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_rsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-023-4"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_dsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-023-3"

  location = detector.locations[_]

  item := location.filename
}

# openssl pkey rsa encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in ["CR-023-7", "CR-023-8"]

    data_type = detector.data_types[_]
    data_type.uuid == password_uuid

    location = data_type.locations[_]
    location.filename in openssl_pkey_rsa_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": data.bearer.common.severity_of_datatype(data_type),
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# openssl pkey dsa encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in ["CR-023-7"]

    data_type = detector.data_types[_]
    data_type.uuid == password_uuid

    location = data_type.locations[_]
    location.filename in openssl_pkey_dsa_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": data.bearer.common.severity_of_datatype(data_type),
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# blowfish encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == "CR-023-6"

    data_type = detector.data_types[_]
    data_type.uuid == password_uuid

    location = data_type.locations[_]
    location.filename in blowfish_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": data.bearer.common.severity_of_datatype(data_type),
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# rc4 encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == "CR-023-1"

    data_type = detector.data_types[_]
    data_type.uuid == password_uuid

    location = data_type.locations[_]
    location.filename in rc4_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": data.bearer.common.severity_of_datatype(data_type),
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == input.policy_id

    data_type = detector.data_types[_]
    data_type.uuid == password_uuid

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
