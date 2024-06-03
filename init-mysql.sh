#!/bin/bash
service mysql stop
mysqld_safe --skip-grant-tables &
sleep 5
mysql -u root <<EOF
USE mysql;
UPDATE user SET authentication_string=PASSWORD('root') WHERE User='root';
FLUSH PRIVILEGES;
EOF
killall mysqld
service mysql start
