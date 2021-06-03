#!/bin/bash

case $1 in
   on)
        systemctl disable pitally
        systemctl stop hostapd
        systemctl stop dnsmasq
        systemctl stop dhcpcd
        cp /usr/local/etc/ap/enabled/dhcpcd.conf /etc/dhcpcd.conf
        cp /usr/local/etc/ap/enabled/hostapd.conf /etc/hostapd/hostapd.conf
        cp /usr/local/etc/ap/enabled/dnsmasq.conf /etc/dnsmasq.conf
        systemctl daemon-restart
        systemctl enable hostapd
        systemctl start hostapd
        systemctl start dnsmasq
        systemctl start dhcpcd
  ;;

   off)
        systemctl stop hostapd
        systemctl disable hostapd
        systemctl stop dnsmasq
        systemctl stop dhcpcd
        cp /usr/local/etc/ap/disabled/dhcpcd.conf /etc/dhcpcd.conf
        cp /usr/local/etc/ap/disabled/hostapd.conf /etc/hostapd/hostapd.conf
        cp /usr/local/etc/ap/disabled/dnsmasq.conf /etc/dnsmasq.conf
        systemctl daemon-restart
        systemctl start dnsmasq
        systemctl start dhcpcd
        systemctl enable pitally
        systemctl start pitally
   ;;
esac
