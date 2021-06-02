<?php

function update_wpa($ssid,$password,$type="WPA-PSK") {
    $path = "/etc/wpa_supplicant/";
    $filename = "wpa_supplicant.conf";
    $file = file($path . $filename);

    //go through the file and collect everything up to the first network={ line
    $header = array();
    foreach ($file as $line) {
        if (!preg_match('/^network=\{/',$line)) {
            $header[] = $line;
        } else {
            break;
        }
    }

    //iterate through file array and pull out the network chunks
    $i =0;
    $networks = array();
    foreach ($file as $line) {

        if (preg_match('/^network=\{/',$line)) {
            //capture through to next }
            $p = $i;

            while (!preg_match('/^\}/',$file[$p])) {
                if (!preg_match('/^network=\{/',$file[$p])) {
                    $elements = array_map('trim',explode('=',$file[$p]));
                    $key = $elements[0];
                    $value = $elements[1];
                    $networks[$i][$key] = $value;
                }
                $p++;

            }

        }
        $i++;
    }

    //check to see if we have a network with the same SSID as provided.  If so discard
    //the old config
    foreach ($networks as $key=>$values) {
        if ($networks[$key]['ssid']  == '"' .$ssid . '"') {
            unset($networks[$key]);
        }
    }

    //add in the new network into the array
    $networks[] = array('ssid' => '"' . $ssid . '"',
                    'psk' => '"' . $password . '"',
                    'key_mgmt' => $type);


    //assemble the new wpa_supplicant file in an array
    $new_wpa = array();

    foreach ($header as $line) {
        $new_wpa[] = $line;
    }

    //add in the networks
    foreach($networks as $network) {
        $new_wpa[] = "network={" . "\n";
        $new_wpa[] = '	ssid=' . $network['ssid'] . "\n";
        $new_wpa[] = '	psk=' . $network['psk'] . "\n";
        $new_wpa[] = '	key_mgmt=' .$network['key_mgmt'] . "\n";
        $new_wpa[] = '}' . "\n";
        $new_wpa[] = "" . "\n";
    }

    //backup old file
    exec('sudo cp ' .$path . $filename . ' ' . $path . $filename . '.' . time());

    //write the new file
    file_put_contents($filename,$new_wpa);
    exec('sudo mv ' . $filename . ' ' . $path . $filename);

}

function update_tally ($api_ip,$tally_action,$tally_value) {
    //updates the configuration for pitally.py
    $path = '/usr/local/etc/pitally/';
    $file = 'tally_config.json';

    $config = array('ip'=>$api_ip,
                    'tally_action' => $tally_action,
                    'tally_value' => strtoupper($tally_value));

    $config_json = json_encode($config) . "\n";

    //backup current config file
    rename($path . $file,$path . $file . "." .time());

    //write new file
    file_put_contents($path . $file,$config_json);
}
