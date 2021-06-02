#!/bin/bash

case $1 in
   on)
	cp /usr/local/etc/ap/enabled/dhcpcd.conf /etc/dhcpcd.conf
	cp /usr/local/etc/ap/enabled/hostapd.conf /etc/hostapd/hostapd.conf
	cp /usr/local/etc/ap/enabled/dnsmasq.conf /etc/dnsmasq.conf
   ;;

   off)
        cp /usr/local/etc/ap/disabled/dhcpcd.conf /etc/dhcpcd.conf
        cp /usr/local/etc/ap/disabled/hostapd.conf /etc/hostapd/hostapd.conf
        cp /usr/local/etc/ap/disabled/dnsmasq.conf /etc/dnsmasq.conf
   ;;
esac
