#!/bin/bash

case $1 in
   on)
	cp /usr/local/etc/ap/enabled/dhcpcd.conf /etc/dhcpcd.conf
	cp /usr/local/etc/ap/enabled/interfaces /etc/network/interfaces
	cp /usr/local/etc/ap/enabled/hostapd.conf /etc/hostapd/hostapd.conf
	cp /usr/local/etc/ap/enabled/hostapd /etc/default/hostapd
	cp /usr/local/etc/ap/enabled/dnsmasq.conf /etc/dnsmasq.conf
   ;;

   off)
        cp /usr/local/etc/ap/disabled/dhcpcd.conf /etc/dhcpcd.conf
        cp /usr/local/etc/ap/disabled/interfaces /etc/network/interfaces
        cp /usr/local/etc/ap/disabled/hostapd.conf /etc/hostapd/hostapd.conf
        cp /usr/local/etc/ap/disabled/hostapd /etc/default/hostapd
        cp /usr/local/etc/ap/disabled/dnsmasq.conf /etc/dnsmasq.conf
   ;;
esac
