Group: BytemanD
Name: stackcrud
Version: VERSION
Release: 1
Summary: Golang OpenStack Client
License: ASL 2.0

Source0: stackcrud
Source1: stackcrud-template.yaml

%global CONFIG_DIRNAME stackcrud
%global CONFIG_PATH /etc/${CONFIG_DIRNAME}

%description
Golang EC Tools


%prep
#cp -p %SOURCE0 %{_builddir}
mkdir -p %{_builddir}${CONFIG_PATH}


%files
%{_bindir}/stackcrud
%{_sysconfdir}/stackcrud/stackcrud-template.yaml

%install
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME}

install -p -m 755 -t %{buildroot}%{_bindir} %{SOURCE0}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE1}

%post

cd %{_sysconfdir}/stackcrud/
if [[ ! -f stackcrud.yaml ]]; then
    cp stackcrud-template.yaml stackcrud.yaml
fi
