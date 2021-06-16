#!/bin/bash

if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit
fi

case $1 in
on)
  systemctl stop pitally
  cp /usr/local/etc/ap/enabled/dnsmasq.conf /etc/dnsmasq.conf
  cp /usr/local/etc/ap/enabled/dhcpcd.conf /etc/dhcpcd.conf
  cp /usr/local/etc/ap/enabled/hostapd.conf /etc/hostapd/hostapd.conf

  systemctl daemon-reload
  systemctl start dnsmasq
  systemctl restart dhcpcd
  systemctl unmask hostapd
  systemctl start hostapd
  /usr/local/bin/ledscli green on
  sleep 2s
  /usr/local/bin/ledscli yellow on
  sleep 2s
  /usr/local/bin/ledscli red on
  ;;

off)
  systemctl stop hostapd
  systemctl mask hostapd
  systemctl stop dnsmasq
  cp /usr/local/etc/ap/disabled/dhcpcd.conf /etc/dhcpcd.conf
  cp /usr/local/etc/ap/disabled/hostapd.conf /etc/hostapd/hostapd.conf
  cp /usr/local/etc/ap/disabled/dnsmasq.conf /etc/dnsmasq.conf
  systemctl daemon-reload
  systemctl restart dhcpcd
  systemctl start pitally
  ;;
esac
