package autodiscovery_test

func Test() {
	lsofNetworkOutput := `p75553
	f4
	n*:8080
	f5
	n*:*`

	lsofFileOutput := `p75553
	fcwd
	n/
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/bin/httpd
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_mpm_prefork.so
	ftxt
	n/opt/homebrew/Cellar/apr-util/1.6.3/lib/libaprutil-1.0.dylib
	ftxt
	n/opt/homebrew/Cellar/pcre2/10.42/lib/libpcre2-8.0.dylib
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authn_file.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authn_core.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authz_host.so
	ftxt
	n/opt/homebrew/Cellar/apr/1.7.2/lib/libapr-1.0.dylib
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authz_groupfile.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authz_user.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_authz_core.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_access_compat.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_auth_basic.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_reqtimeout.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_filter.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_mime.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_log_config.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_env.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_headers.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_setenvif.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_version.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_unixd.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_status.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_dir.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_autoindex.so
	ftxt
	n/opt/homebrew/Cellar/httpd/2.4.56/lib/httpd/modules/mod_alias.so
	f0
	n/dev/null
	f1
	n/dev/null
	f2
	n/opt/homebrew/var/log/httpd/error_log
	f3
	n[ctl com.apple.netsrc id 6 unit 37]
	f4
	n*:http-alt
	f5
	n*:*
	f6
	n->0x76b0ef67c329cc0c
	f7
	n->0x61e75964f19947b5
	f8
	n->0x2bc69703b03483ba
	f9
	n/opt/homebrew/var/log/httpd/access_log
	f10
	ncount=1, state=0x8`
}
