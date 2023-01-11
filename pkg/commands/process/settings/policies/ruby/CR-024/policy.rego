package bearer.cr_024

import data.bearer.common

import future.keywords

password_uuid := "02bb0d3a-2c8c-4842-be1c-c057f0dccd63"

files_with_password_encryption_calls contains item if {
  some detector in input.dataflow.risks
  detector.detector_id in [
    "CR-024-1",
    "CR-024-6",
    "CR-024-7",
    "CR-024-8"
  ]

  data_type = detector.data_types[_]
  data_type.uuid == password_uuid

  location = detector.locations[_]

  item := location.filename
}

rc4_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-024-5"

  location = detector.locations[_]

  item := location.filename
}

blowfish_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-024-2"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_rsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-024-4"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_dsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "CR-024-3"

  location = detector.locations[_]

  item := location.filename
}

# openssl pkey rsa encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in ["CR-024-7", "CR-024-8"]

    data_type = detector.data_types[_]
    data_type.uuid != password_uuid

    location = data_type.locations[_]
    location.filename in openssl_pkey_rsa_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# openssl pkey dsa encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in ["CR-024-7"]

    data_type = detector.data_types[_]
    data_type.uuid != password_uuid

    location = data_type.locations[_]
    location.filename in openssl_pkey_dsa_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# blowfish encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == "CR-024-6"

    data_type = detector.data_types[_]
    data_type.uuid != password_uuid

    location = data_type.locations[_]
    location.filename in blowfish_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

# rc4 encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == "CR-024-1"

    data_type = detector.data_types[_]
    data_type.uuid != password_uuid

    location = data_type.locations[_]
    location.filename in rc4_files

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in [
      "CR-024-2",
      "CR-024-3",
      "CR-024-4",
      "CR-024-5",
    ]

    data_type = detector.data_types[_]
    location = data_type.locations[_]
    # NOT in a file with an encryption method call
    not location in files_with_password_encryption_calls

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}

policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id == "detect_ruby_weak_encryption"

    # NOT password data type
    data_type = detector.data_types[_]
    data_type.uuid != password_uuid

    location = data_type.locations[_]

    item := {
        "category_groups": data.bearer.common.groups_for_datatype(data_type),
        "severity": "low",
        "filename": location.filename,
        "line_number": location.line_number,
        "parent_line_number": location.parent.line_number,
        "parent_content": location.parent.content
    }
}
