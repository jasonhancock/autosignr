%define debug_package %{nil}

Summary:        Certificate Autosigner for Puppet
Name:           autosignr
Version:        %{version}
Release:        1%{?dist}
License:        MIT
Group:          Development/Languages
URL:            https://github.com/jasonhancock/autosignr
Source0:        %{name}-%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
#Requires:       daemonize

%description
A daemon to watch for Puppet CSR's, validate they came from instances managed by us, then sign them.

%prep
%setup -q -n %{name}-%{version}

%build

export GOPATH=$RPM_BUILD_DIR/%{name}-%{version}
cd src/github.com/jasonhancock/autosignr && make

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/sbin
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/bin/autosignr $RPM_BUILD_ROOT/usr/sbin/
install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/bin/autocleanr $RPM_BUILD_ROOT/usr/sbin/

mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/systemd/system/
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/autosignr
install -m 0644 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/autosignr/packaging/redhat/autosignr.logrotate $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d/autosignr
install -m 0644 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/autosignr/packaging/redhat/autosignr.service $RPM_BUILD_ROOT/%{_sysconfdir}/systemd/system/autosignr.service
install -m 0644 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/autosignr/config.yaml $RPM_BUILD_ROOT/%{_sysconfdir}/autosignr/config.yaml
install -m 0644 $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/jasonhancock/autosignr/autocleanr.yaml $RPM_BUILD_ROOT/%{_sysconfdir}/autosignr/autocleanr.yaml

mkdir -p $RPM_BUILD_ROOT%{_localstatedir}/log/autosignr

%post
systemctl enable autosignr.service

%preun
if [ $1 = 0 ]; then
    systemctl stop autosignr.service > /dev/null 2>&1
    systemctl disable autosignr.service
fi

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/usr/sbin/autosignr
%config(noreplace) %{_sysconfdir}/logrotate.d/autosignr
%{_sysconfdir}/systemd/system/autosignr.service
%config(noreplace) %{_sysconfdir}/autosignr/config.yaml
%config(noreplace) %{_sysconfdir}/autosignr/autocleanr.yaml
%attr(0700,root,root) %dir %{_localstatedir}/log/autosignr
