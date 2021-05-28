<?php
$file = file_get_contents('config.json');

$json = json_decode($file);
print_r($json);
