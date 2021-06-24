#!/bin/bash

rm *.pem

openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=CN/ST=Zhejiang/L=Hangzhou/O=Like/CN=*like.com/emailAddress=cert@like.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text


openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=CN/ST=Zhejiang/L=Hangzhou/O=PCBOOK/OU=Computer/CN=*.pcbook.com/emailAddress=cert@pcbook.com"

openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text


openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=CN/ST=Zhejiang/L=Hangzhou/O=PCClient/OU=Computer/CN=*.pcclient.com/emailAddress=cert@pcclient.com"

openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text
