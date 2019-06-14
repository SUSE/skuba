# Copyright (c) 2019 SUSE LLC.
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

Name:      caasp-test
Version:   1
Release:   1
Summary:   CaaSP test package
License:   Apache-2.0
BuildArch: noarch

%description
CaaSP test package

%prep
# nothing to be done

%build
cat > caasp-test <<EOF
#!/usr/bin/env bash
echo CaaSP test version %{version}
EOF

%install
mkdir -p %{buildroot}/usr/bin/
install -m 755 caasp-test %{buildroot}/usr/bin/caasp-test

%files
/usr/bin/caasp-test

%changelog
