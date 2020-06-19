  - [ SUSEConnect, -r, ${caasp_registry_code} ]
  - [ SUSEConnect, -p, sle-module-containers/15.1/x86_64 ]
  - [ SUSEConnect, -p, caasp/5/x86_64, -r, ${caasp_registry_code} ]
%{ if ha_registry_code != "" ~}
  - [ SUSEConnect, -p, sle-ha/15.1/x86_64, -r, ${ha_registry_code} ]
%{ endif ~}
