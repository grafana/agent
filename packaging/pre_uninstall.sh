#!/bin/sh

set -e

[ -f /etc/sysconfig/grafana-agent ] && . /etc/sysconfig/grafana-agent

startAgent() {
  if [ -x /bin/systemctl ] ; then
    /bin/systemctl daemon-reload
		/bin/systemctl start grafana-agent.service
	elif [ -x /etc/init.d/grafana-agent ] ; then
		/etc/init.d/grafana-agent start
	elif [ -x /etc/rc.d/init.d/grafana-agent ] ; then
		/etc/rc.d/init.d/grafana-agent start
	fi
}

stopAgent() {
	if [ -x /bin/systemctl ] ; then
		/bin/systemctl stop grafana-agent.service > /dev/null 2>&1 || :
	elif [ -x /etc/init.d/grafana-agent ] ; then
		/etc/init.d/grafana-agent stop
	elif [ -x /etc/rc.d/init.d/grafana-agent ] ; then
		/etc/rc.d/init.d/grafana-agent stop
	fi
}

# final uninstallation $1=0 
# If other copies of this RPM are installed, then $1>0 

if [ $1 -eq 0 ] ; then
	stopAgent
fi
exit 0
