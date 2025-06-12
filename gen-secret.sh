#!/bin/bash
if ! command -v openssl &> /dev/null; then
	echo "OpenSSL package is not installed."
	exit 1	
fi

openssl rand -hex 32
