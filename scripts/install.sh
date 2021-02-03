#! /bin/bash

installSoftware() {
    apt -qq -y install nginx default-mysql-client
}

installMyAccounts() {
    mkdir -p /var/www/myaccounts
    curl -Lo- https://github.com/sunshineplan/myaccounts/releases/download/v1.0/release.tar.gz | tar zxC /var/www/myaccounts
    cd /var/www/myaccounts
    chmod +x myaccounts
}

configMyAccounts() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /run/myaccounts.sock): ' unix
    [ -z $unix ] && unix=/run/myaccounts.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/myaccounts.log): ' log
    [ -z $log ] && log=/var/log/app/myaccounts.log
    read -p 'Please enter update URL: ' update
    read -p 'Please enter exclude files: ' exclude
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/myaccounts/config.ini.default > /var/www/myaccounts/config.ini
    sed -i "s/\$header/$header/" /var/www/myaccounts/config.ini
    sed -i "s/\$value/$value/" /var/www/myaccounts/config.ini
    sed -i "s/\$domain/$domain/" /var/www/myaccounts/config.ini
    sed -i "s,\$unix,$unix," /var/www/myaccounts/config.ini
    sed -i "s,\$log,$log," /var/www/myaccounts/config.ini
    sed -i "s/\$host/$host/" /var/www/myaccounts/config.ini
    sed -i "s/\$port/$port/" /var/www/myaccounts/config.ini
    sed -i "s,\$update,$update," /var/www/myaccounts/config.ini
    sed -i "s|\$exclude|$exclude|" /var/www/myaccounts/config.ini
    ./myaccounts install
    service myaccounts start
}

writeLogrotateScrip() {
    if [ ! -f '/etc/logrotate.d/app' ]; then
	cat >/etc/logrotate.d/app <<-EOF
		/var/log/app/*.log {
		    copytruncate
		    rotate 12
		    compress
		    delaycompress
		    missingok
		    notifempty
		}
		EOF
    fi
}

createCronTask() {
    cp -s /var/www/myaccounts/scripts/myaccounts.cron /etc/cron.monthly/myaccounts
    chmod +x /var/www/myaccounts/scripts/myaccounts.cron
}

setupNGINX() {
    cp -s /var/www/myaccounts/scripts/myaccounts.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/myaccounts/scripts/myaccounts.conf
    sed -i "s,\$unix,$unix," /var/www/myaccounts/scripts/myaccounts.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installMyAccounts
    configMyAccounts
    writeLogrotateScrip
    createCronTask
    setupNGINX
}

main