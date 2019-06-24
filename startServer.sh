#!/bin/sh
echo "Insira o numero de instâncias a serem criadas: "
read inst;

var=$((inst+8000))
cd server 

echo "upstream chatgo {" > /etc/nginx/configServer/serversgo.conf
for ((i=8000; i < var; i++))
do
        echo "server localhost:"$i";" >> /etc/nginx/configServer/serversgo.conf  
done

echo "}" >> /etc/nginx/configServer/serversgo.conf
systemctl restart nginx

for ((i=8000; i < var; i++))
do
    teste=$((var-1))
    if [ $i -eq $teste ]
    then
        ./server -port :$i
    else
        ./server -port :$i & \
    fi 
done
