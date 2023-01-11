package bearer.weak_encryption_library

import data.bearer.common
import data.bearer.encryption

import future.keywords

password_uuid := "02bb0d3a-2c8c-4842-be1c-c057f0dccd63"

# openssl pkey rsa encryption
policy_failure contains item if {
    some detector in input.dataflow.risks
    detector.detector_id in ["ruby_openssl_pkey_rsa", "ruby_openssl_pkey"]

    data_type = detector.data_types[_]
    data_type.uuid != data.bearer.encryption_common.password_uuid

    location = data_type.locations[_]
    location.filename in data.bearer.encryption_common.openssl_pkey_rsa_files

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
    detector.detector_id in ["ruby_openssl_pkey_dsa", "ruby_openssl_pkey"]

    data_type = detector.data_types[_]
    data_type.uuid != data.bearer.encryption_common.password_uuid

    location = data_type.locations[_]
    location.filename in data.bearer.encryption_common.openssl_pkey_dsa_files

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
    detector.detector_id == "ruby_blowfish"

    data_type = detector.data_types[_]
    data_type.uuid != data.bearer.encryption_common.password_uuid

    location = data_type.locations[_]
    location.filename in data.bearer.encryption_common.blowfish_files

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
    detector.detector_id == "ruby_encrypt"

    data_type = detector.data_types[_]
    data_type.uuid != data.bearer.encryption_common.password_uuid

    location = data_type.locations[_]
    location.filename in data.bearer.encryption_common.rc4_files

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
      "ruby_initialize_rc4_encryption",
      "ruby_initialize_blowfish_encryption",
      "ruby_initialize_openssl_pkey_rsa_encryption",
      "ruby_initialize_openssl_pkey_dsa_encryption",
    ]

    data_type = detector.data_types[_]
    location = data_type.locations[_]
    # NOT in a file with an encryption method call
    not location in data.bearer.encryption_common.files_with_password_encryption_calls

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
    data_type.uuid != data.bearer.encryption_common.password_uuid

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
