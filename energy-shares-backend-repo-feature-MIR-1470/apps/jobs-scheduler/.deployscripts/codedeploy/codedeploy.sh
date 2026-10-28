#!/bin/bash

#systemctl
OLD_SYSTEMCTL_CONFIG_PATH=/etc/systemd/system/energyshares-jobs-scheduler.service
NEW_SYSTEMCTL_CONFIG_PATH=/tmp/build/.deployscripts/systemctl/energyshares-jobs-scheduler.service
SYSTEMCTL_NAME=energyshares-jobs-scheduler
function systemctl_deploy(){
    echo "systemctl_deploy()"
    sudo cp $NEW_SYSTEMCTL_CONFIG_PATH $OLD_SYSTEMCTL_CONFIG_PATH
    sudo systemctl daemon-reload
    sudo systemctl enable $SYSTEMCTL_NAME
    sudo systemctl start $SYSTEMCTL_NAME
}
if [ -f "$OLD_SYSTEMCTL_CONFIG_PATH" ] ; then
    echo "systemctl file exist"
    diff $OLD_SYSTEMCTL_CONFIG_PATH $NEW_SYSTEMCTL_CONFIG_PATH
    DIFF_RESULT=$?
    echo "systemctl diff a b result :  ${DIFF_RESULT}"
    if [ ${DIFF_RESULT} -eq "0" ]
    then
        echo "systemctl OLD file and NEW file are same"
    else
        echo "systemctl OLD file and NEW file are different"
        systemctl_deploy
    fi
else
    echo "systemctl file not exist"
    systemctl_deploy
fi

#syslog
OLD_SYSLOG_CONFIG_PATH=/etc/rsyslog.d/energyshares-jobs-scheduler.conf
NEW_SYSLOG_CONFIG_PATH=/tmp/build/.deployscripts/syslog/energyshares-jobs-scheduler.conf
function syslog_deploy(){
    echo "syslog_deploy()"
    sudo cp $NEW_SYSLOG_CONFIG_PATH $OLD_SYSLOG_CONFIG_PATH
    sudo systemctl restart rsyslog
}
if [ -f "$OLD_SYSLOG_CONFIG_PATH" ] ; then
    echo "syslog file exist"
    diff $OLD_SYSLOG_CONFIG_PATH $NEW_SYSLOG_CONFIG_PATH
    DIFF_RESULT=$?
    echo "syslog diff a b result :  ${DIFF_RESULT}"
    if [ ${DIFF_RESULT} -eq "0" ]
    then
        echo "syslog OLD file and NEW file are same"
    else
        echo "syslog OLD file and NEW file are different"
        syslog_deploy
    fi
else
    echo "syslog file not exist"
    syslog_deploy
fi

#logrotate
OLD_LOGROTATE_CONFIG_PATH=/etc/logrotate.d/energyshares-jobs-scheduler-service
NEW_LOGROTATE_CONFIG_PATH=/tmp/build/.deployscripts/logrotate/energyshares-jobs-scheduler-service
function logrotate_deploy(){
    echo "logrotate_deploy()"
    sudo cp $NEW_LOGROTATE_CONFIG_PATH $OLD_LOGROTATE_CONFIG_PATH
    sudo logrotate -d $OLD_LOGROTATE_CONFIG_PATH
}
if [ -f "$OLD_LOGROTATE_CONFIG_PATH" ] ; then
    echo "logrotate file exist"
    diff $OLD_LOGROTATE_CONFIG_PATH $NEW_LOGROTATE_CONFIG_PATH
    DIFF_RESULT=$?
    echo "logrotate diff a b result :  ${DIFF_RESULT}"
    if [ ${DIFF_RESULT} -eq "0" ]
    then
        echo "logrotate OLD file and NEW file are same"
    else
        echo "logrotate OLD file and NEW file are different"
        logrotate_deploy
    fi
else
    echo "logrotate file not exist"
    logrotate_deploy
fi


#service
OLD_SERVICE_CONFIG_PATH=/usr/local/bin/energyshares-jobs-scheduler-service
NEW_SERVICE_CONFIG_PATH=/tmp/build/.deployscripts/service/energyshares-jobs-scheduler-service
function service_deploy(){
    echo "service_deploy()"
    sudo cp $NEW_SERVICE_CONFIG_PATH $OLD_SERVICE_CONFIG_PATH
    sudo chmod 777 $OLD_SERVICE_CONFIG_PATH
}
if [ -f "$OLD_SERVICE_CONFIG_PATH" ] ; then
    echo "service file exist"
    diff $OLD_SERVICE_CONFIG_PATH $NEW_SERVICE_CONFIG_PATH
    DIFF_RESULT=$?
    echo "service diff a b result :  ${DIFF_RESULT}"
    if [ ${DIFF_RESULT} -eq "0" ]
    then
        echo "service OLD file and NEW file are same"
    else
        echo "service OLD file and NEW file are different"
        service_deploy
    fi
else
    echo "service file not exist"
    service_deploy
fi

#promtail-config update
OLD_PROMTAIL_CONFIG_PATH=/home/ubuntu/promtail/config/promtail-config.yaml
NEW_PROMTAIL_CONFIG_PATH=/tmp/build/.deployscripts/promtail/promtail-config.yaml
function promtail_deploy(){
	echo "promtail_deploy()"
	sudo cp $NEW_PROMTAIL_CONFIG_PATH $OLD_PROMTAIL_CONFIG_PATH 
	sudo systemctl restart promtail
}
if [ -f "$OLD_PROMTAIL_CONFIG_PATH" ] ; then
	echo "promtail file exist"
	diff $OLD_PROMTAIL_CONFIG_PATH $NEW_PROMTAIL_CONFIG_PATH
	DIFF_RESULT=$?
	echo "promtail diff a b result :  ${DIFF_RESULT}"
	if [ ${DIFF_RESULT} -eq "0" ]
	then
		echo "promtail OLD file and NEW file are same"
	else
		echo "promtail OLD file and NEW file are different"
		promtail_deploy
	fi
else
	echo "promtail file not exist"
	promtail_deploy
fi

#binary

sudo mv /tmp/build /home/ubuntu/bin-new
sudo chown -R ubuntu:ubuntu /home/ubuntu/bin-new
sudo mv /home/ubuntu/bin /home/ubuntu/bin_$(date +"%Y%m%d%H%M")
sudo mv /home/ubuntu/bin-new /home/ubuntu/bin

sudo systemctl stop energyshares-jobs-scheduler.service
sudo setcap CAP_NET_BIND_SERVICE=+eip /home/ubuntu/bin/jobs-scheduler
sudo systemctl start energyshares-jobs-scheduler.service
