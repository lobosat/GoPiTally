<html>
<head>
    <link href="form.css" rel="stylesheet">
</head>
<body>

<?php
if ($_POST['reboot'] == 'Reboot') {
    exec('sudo /usr/local/bin/doap.sh off');
    exec('sudo /usr/sbin/reboot');
}

if ($_POST['accept'] == 'Accept') {
    //Apply Settings
    require_once('lib/update.php');
    $api_ip = $_POST['api_ip'];
    $tally_action = $_POST['tally_action'];
    $tally_value = $_POST['tally_value'];
    $ssid = $_POST['ssid'];
    $wifi_password = $_POST['wifi_password'];

    update_tally($api_ip,$tally_action,$tally_value);
    update_wpa($ssid,$wifi_password);

?>
    <div id="main">
        <div id="first" style="width: 400px; height: 300px">
            <h4>Confirm Settings</h4>
            <form method="post">
                <h4>Settings Saved</h4>
                <br/>
                <input type="submit" name="reboot" id="reboot" value="Reboot"/>
        </div>
    </div>
<?php
}

  if ($_POST['sub_form'] == 'Submit') {
      //display confirmation page
?>
<div id="main">
<div id="first" style="width: 400px; height: 300px">
    <h4>Confirm Settings</h4>
    <form method="post">
    <table>
        <tr>
            <td>vMix API IP address</td>
            <td><?php echo $_POST['api_ip']?></td>
        </tr>
        <tr>
            <td>Tally Action</td>
            <td><?php echo $_POST['tally_action']?></td>
        </tr>
        <tr>
            <td>Tally Value</td>
            <td><?php echo $_POST['tally_value']?></td>
        </tr>
        <tr>
            <td>WiFi SSID</td>
            <td><?php echo $_POST['ssid']?></td>
        </tr>
        <tr>
            <td>WiFi Password</td>
            <td><?php echo $_POST['wifi_password']?></td>
        </tr>
        <tr>

            <td><input type="submit" name="redo" id="redo" value="Redo"\> </td>
            <td><input type="submit" name="accept" id="accept" value="Accept" </td>
                <input type="hidden" name="api_ip" id="api_ip" value="<?php echo $_POST['api_ip']?>"/>
                <input type="hidden" name="ssid" id="ssid" value="<?php echo $_POST['ssid']?>"/>
                <input type="hidden" name="wifi_password" id="wifi_password" value="<?php echo $_POST['wifi_password']?>"/>
                <input type="hidden" name="tally_action" id="tally_action" value="<?php echo $_POST['tally_action']?>"/>
                <input type="hidden" name="tally_value" id="tally_value" value="<?php echo $_POST['tally_value']?>"/>

        </tr>
    </table>
    </form>
    </div>
    </div>
<?php
  }

if ( empty($_POST['sub_form']) && empty($_POST['accept']) && empty($_POST['reboot'])){
?>
<div id="main">
    <div id="first">
        <h4>PiTally Configuration Page</h4>
        <form method="post">

            <label for="api_ip">vMix API IP address</label>
            <input type="text" name="api_ip" id="api_ip"/>
            <br/>
            <label for="tally_action">Tally Action</label>
            <select name="tally_action" id="tally_action">
                <option value="input">Input</option>
                <option value="bus">Bus</option>
            </select>
            <br/>
            <label for="tally_value">Tally Value</label>
            <input type="text" name="tally_value" id="tally_value" size=3/>
            <br/>
            <label for="ssid">WiFi SSID</label>
            <input type="text" name="ssid" id="ssid"/>
            <br/>
            <label for="wifi_password">WiFi Password</label>
            <input type="text" name="wifi_password" id="wifi_password"/>
            <br/><br/>
            <input type="submit" name="sub_form" id="sub_form" value="Submit"/>
        </form>
     </div>
 </div>
</body>
</html>
<?php } ?>