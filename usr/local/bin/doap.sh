#!/bin/bash

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
