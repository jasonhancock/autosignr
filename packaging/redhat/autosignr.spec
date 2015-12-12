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

mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/rc.d/init.d
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d
mkdir -p $RPM_BUILD_ROOT/%{_sysconfdir}/sysconfig
#install -m 0755 $RPM_BUILD_DIR/%{name}-%{version}/packaging/redhat/amproxy.logrotate $RPM_BUILD_ROOT/%{_sysconfdir}/logrotate.d/amproxy

#mkdir -p $RPM_BUILD_ROOT%{_localstatedir}/log/amproxy

%post
/sbin/chkconfig --add autosignr

%preun
if [ $1 = 0 ]; then
    /sbin/service autosignr stop > /dev/null 2>&1
    /sbin/chkconfig --del autosignr
fi

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
/usr/sbin/autosignr
#%{_sysconfdir}/rc.d/init.d/amproxy
#%config(noreplace) %{_sysconfdir}/logrotate.d/amproxy

#%attr(0700,amproxy,amproxy) %dir %{_localstatedir}/log/amproxy
