Group: BytemanD
Name: skyman
Version: VERSION
Release: 1
Summary: Golang OpenStack Client
License: ASL 2.0

Source0: skyman
Source1: skyman-template.yaml
Source2: zh_CN.toml

%global CONFIG_DIRNAME skyman
%global CONFIG_PATH /etc/${CONFIG_DIRNAME}
%global SHARE_LOCALE_PATH /usr/share/%{CONFIG_DIRNAME}/locale

%description
Golang EC Tools


%prep
#cp -p %SOURCE0 %{_builddir}
mkdir -p %{_builddir}${CONFIG_PATH}
mkdir -p %{_builddir}${SHARE_LOCALE_PATH}


%files
%{_bindir}/skyman
%{_sysconfdir}/skyman/skyman-template.yaml
%{SHARE_LOCALE_PATH}/zh_CN.toml

%install
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME}
install -m 755 -d %{buildroot}%{SHARE_LOCALE_PATH}

install -p -m 755 -t %{buildroot}%{_bindir} %{SOURCE0}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE1}
install -p -m 755 -t %{buildroot}%{SHARE_LOCALE_PATH} %{SOURCE2}

%post

cd %{_sysconfdir}/skyman/
if [[ ! -f skyman.yaml ]]; then
    cp skyman-template.yaml skyman.yaml
fi
