#! /bin/bash

installSoftware() {
    apt -qq -y install nginx default-mysql-client
}

installAuth() {
    mkdir -p /var/www/auth
    curl -Lo- https://github.com/sunshineplan/auth/releases/download/v1.0/release.tar.gz | tar zxC /var/www/auth
    cd /var/www/auth
    chmod +x auth
}

configMyAuth() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter auth server domain: ' domain
    read -p 'Please enter unix socket(default: /run/auth.sock): ' unix
    [ -z $unix ] && unix=/run/auth.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/auth.log): ' log
    [ -z $log ] && log=/var/log/app/auth.log
    read -p 'Please enter update URL: ' update
    read -p 'Please enter exclude files: ' exclude
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/auth/config.ini.default > /var/www/auth/config.ini
    sed -i "s/\$header/$header/" /var/www/auth/config.ini
    sed -i "s/\$value/$value/" /var/www/auth/config.ini
    sed -i "s,\$domain,$domain," /var/www/auth/config.ini
    sed -i "s,\$unix,$unix," /var/www/auth/config.ini
    sed -i "s,\$log,$log," /var/www/auth/config.ini
    sed -i "s/\$host/$host/" /var/www/auth/config.ini
    sed -i "s/\$port/$port/" /var/www/auth/config.ini
    sed -i "s,\$update,$update," /var/www/auth/config.ini
    sed -i "s|\$exclude|$exclude|" /var/www/auth/config.ini
    ./auth install
    service auth start
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
    cp -s /var/www/auth/scripts/auth.cron /etc/cron.monthly/auth
    chmod +x /var/www/auth/scripts/auth.cron
}

setupNGINX() {
    cp -s /var/www/auth/scripts/auth.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$server_name/" /var/www/auth/scripts/auth.conf
    sed -i "s,\$unix,$unix," /var/www/auth/scripts/auth.conf
    service nginx reload
}

main() {
    read -p 'Please enter server name:' server_name
    installSoftware
    installAuth
    configAuth
    writeLogrotateScrip
    createCronTask
    setupNGINX
}

main