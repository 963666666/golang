#!/bin/bash

###################################################
# set
redisAddr="147.139.1.191:6379"
mysqlAddr="mrpoker:c07XCLzBMp82DUe_9foWgr@tcp(147.139.1.191:3306)\/mrpoker?charset=utf8mb4"
mqAddr="amqp:\/\/guest:guest@147.139.1.191:5672"
influxdbAddr="http:\/\/147.139.1.191:8086"

output=output
###################################################

cd "$(dirname "$0")"
pwd

###################################################
# update config
redis=`cat server.conf|awk -F"[redis_addr]" '/redis_addr/{print $x}'`

mysql=`cat server.conf|awk -F"[mysql_addr]" '/mysql_addr/{print $x}'`

mq=`cat server.conf|awk -F"[mq_addr]" '/mq_addr/{print $x}'`

influxdb=`cat server.conf|awk -F"[influxdb_addr]" '/influxdb_addr/{print $x}'`
#redis
if [ ! -n "$redis" ]; then
    sed -i "/{/a\"redis_addr\":\"$redisAddr\"," server.conf
else
    sed -i "s/\"redis_addr\":.*$/\"redis_addr\":\"$redisAddr\",/" server.conf
fi
#mysql
if [ ! -n "$mysql" ]; then
    sed -i "/{/a\"mysql_addr\":\"$mysqlAddr\"," server.conf
else
    sed -i "s/\"mysql_addr\":.*$/\"mysql_addr\":\"$mysqlAddr\",/" server.conf
fi
#mq
if [ ! -n "$mq" ]; then
    sed -i "/{/a\"mq_addr\":\"$mqAddr\"," server.conf
else
    sed -i "s/\"mq_addr\":.*$/\"mq_addr\":\"$mqAddr\",/" server.conf
fi
#influxdb
if [ ! -n "$influxdb" ]; then
    sed -i "/{/a\"influxdb_addr\":\"$influxdbAddr\"," server.conf
else
    sed -i "s/\"influxdb_addr\":.*$/\"influxdb_addr\":\"$influxdbAddr\",/" server.conf
fi

###################################################

rm -rf $output
mkdir $output

go build -a -o $output/server server.go && \
cp server.conf $output && \
cp -rf LuaPoker $output
