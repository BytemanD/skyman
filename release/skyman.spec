Group: BytemanD
Name: skyman
Version: VERSION
Release: 1
Summary: Golang OpenStack Client
License: ASL 2.0

Source0: skyman
Source1: clouds-sample.yaml
Source2: zh_CN.toml
Source3: index.html
Source4: resource-template.yaml
Source5: server-actions-test-template.yaml

%global CONFIG_DIRNAME skyman
%global CONFIG_PATH /etc/${CONFIG_DIRNAME}
%global SHARE_LOCALE_PATH /usr/share/%{CONFIG_DIRNAME}/locale
%global SHARE_STATIC_PATH /usr/share/%{CONFIG_DIRNAME}/static

%description
Golang OpenStack Client


%prep
#cp -p %SOURCE0 %{_builddir}
mkdir -p %{_builddir}${CONFIG_PATH}
mkdir -p %{_builddir}${SHARE_LOCALE_PATH}
mkdir -p %{_builddir}${SHARE_STATIC_PATH}


%files
%{_bindir}/skyman
%{_sysconfdir}/skyman/clouds-sample.yaml
%{_sysconfdir}/skyman/resource-template.yaml
%{_sysconfdir}/skyman/server-actions-test-template.yaml
%{SHARE_LOCALE_PATH}/zh_CN.toml
%{SHARE_STATIC_PATH}/index.html

%install
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME}
install -m 755 -d %{buildroot}%{SHARE_LOCALE_PATH}
install -m 755 -d %{buildroot}%{SHARE_STATIC_PATH}

install -p -m 755 -t %{buildroot}%{_bindir} %{SOURCE0}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE1}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE4}
install -p -m 755 -t %{buildroot}%{_sysconfdir}/%{CONFIG_DIRNAME} %{SOURCE5}
install -p -m 755 -t %{buildroot}%{SHARE_LOCALE_PATH} %{SOURCE2}
install -p -m 755 -t %{buildroot}%{SHARE_STATIC_PATH} %{SOURCE3}

%post

cd %{_sysconfdir}/skyman/
if [[ ! -f clouds.yaml ]]; then
    cp clouds-sample.yaml clouds.yaml
fi
