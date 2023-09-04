Group: BytemanD
Name: stackcrud
Version: VERSION
Release: 1
Summary: Golang OpenStack Client
License: ASL 2.0

Source0: stackcrud
Source1: stackcrud-template.yaml
Source2: zh_CN.toml

%global CONFIG_DIRNAME stackcrud
%global CONFIG_PATH /etc/${CONFIG_DIRNAME}
%global SHARE_LOCALE_PATH /usr/share/%{CONFIG_DIRNAME}/locale

%description
Golang EC Tools


%prep
#cp -p %SOURCE0 %{_builddir}
mkdir -p %{_builddir}${CONFIG_PATH}
mkdir -p %{_builddir}${SHARE_LOCALE_PATH}


%files
%{_bindir}/stackcrud
%{_sysconfdir}/stackcrud/stackcrud-template.yaml
%{SHARE_LOCALE_PATH}/zh_CN.toml

%install
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME}
install -m 755 -d %{buildroot}%{SHARE_LOCALE_PATH}

install -p -m 755 -t %{buildroot}%{_bindir} %{SOURCE0}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE1}
install -p -m 755 -t %{buildroot}%{SHARE_LOCALE_PATH} %{SOURCE2}

%post

cd %{_sysconfdir}/stackcrud/
if [[ ! -f stackcrud.yaml ]]; then
    cp stackcrud-template.yaml stackcrud.yaml
fi
