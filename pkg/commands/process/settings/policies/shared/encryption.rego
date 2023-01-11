package bearer.encryption_common

import future.keywords

password_uuid := "02bb0d3a-2c8c-4842-be1c-c057f0dccd63"

files_with_password_encryption_calls contains item if {
  some detector in input.dataflow.risks
  detector.detector_id in [
    "ruby_openssl_pkey_rsa",
    "ruby_openssl_pkey",
    "ruby_blowfish",
    "ruby_encrypt"
  ]

  data_type = detector.data_types[_]
  data_type.uuid == password_uuid

  location = detector.locations[_]

  item := location.filename
}

rc4_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "ruby_initialize_rc4_encryption"

  location = detector.locations[_]

  item := location.filename
}

blowfish_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "ruby_initialize_blowfish_encryption"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_rsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "ruby_initialize_openssl_pkey_rsa_encryption"

  location = detector.locations[_]

  item := location.filename
}

openssl_pkey_dsa_files contains item if {
  some detector in input.dataflow.risks
  detector.detector_id == "ruby_initialize_openssl_pkey_dsa_encryption"

  location = detector.locations[_]

  item := location.filename
}