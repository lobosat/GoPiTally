<html>
<head>
    <link href="form.css" rel="stylesheet">
</head>
<body>

<?php
if ($_POST['reboot'] == 'Reboot') {
    exec('sudo /usr/local/bin/doap.sh off');
}

if ($_POST['accept'] == 'Accept') {
    //Apply Settings
    require_once('lib/update.php');
    update_tally($_POST);
    update_wpa($_POST['ssid'],$_POST['wifi_password']);

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

            <td>Red Tally</td>
            <td>Type: <?php echo $_POST['red_type']?><br/>
            Value: <?php echo $_POST['red_value']?>
            </td>
        </tr>
        <tr>
            <td>Yellow Tally</td>
            <td>Type: <?php echo $_POST['yellow_type']?><br/>
            Value: <?php echo $_POST['yellow_value']?>
            </td>
        </tr>
        <tr>
            <td>Green Tally</td>
            <td>Type: <?php echo $_POST['green_type']?><br/>
            Value: <?php echo $_POST['green_value']?>
            </td>
        </tr>
        <tr>
            <td>Tally Value</td>
            <td><?php echo $_POST['tally_value']?></td>
        </tr>
        <tr>
            <td>vMix API IP address</td>
            <td><?php echo $_POST['api_ip']?></td>
        </tr>
        <tr>
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


        </tr>
    </table>
    <input type="hidden" name="api_ip" id="api_ip" value="<?php echo $_POST['api_ip']?>"/>
    <input type="hidden" name="ssid" id="ssid" value="<?php echo $_POST['ssid']?>"/>
    <input type="hidden" name="wifi_password" id="wifi_password" value="<?php echo $_POST['wifi_password']?>"/>
    <input type="hidden" name="red_type" id="red_type" value="<?php echo $_POST['red_type']?>"/>
    <input type="hidden" name="red_value" id="red_value" value="<?php echo $_POST['red_value']?>"/>
    <input type="hidden" name="yellow_type" id="yellow_type" value="<?php echo $_POST['yellow_type']?>"/>
    <input type="hidden" name="yellow_value" id="yellow_value" value="<?php echo $_POST['yellow_value']?>"/>
    <input type="hidden" name="green_type" id="green_type" value="<?php echo $_POST['green_type']?>"/>
    <input type="hidden" name="green_value" id="green_value" value="<?php echo $_POST['green_value']?>"/>
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

            <h2>Red</h2>
            <label for="red_type">Type</label>
            <select name="red_type" id="red_type">
                <option value="None">None</option>
                <option value="Input">Input</option>
                <option value="Bus">Bus</option>
                <option value="Streaming">Streaming</option>
                <option value="Recording">Recording</option>
            </select>
            <br/>
            <label for="red_value">Tally Value</label>
            <input type="text" name="red_value" id="red_value" size=3/>
            <br/>
            <h2>Yellow</h2>
            <label for="yellow_type">Type</label>
            <select name="yellow_type" id="yellow_type">
                <option value="None">None</option>
                <option value="Input">Input</option>
                <option value="Bus">Bus</option>
                <option value="Streaming">Streaming</option>
                <option value="Recording">Recording</option>
            </select>
            <br/>
            <label for="yellow_value">Tally Value</label>
            <input type="text" name="yellow_value" id="yellow_value" size=3/>
            <br/>
            <h2>Green</h2>
            <label for="green_type">Type</label>
            <select name="green_type" id="green_type">
                <option value="None">None</option>
                <option value="Input">Input</option>
                <option value="Bus">Bus</option>
                <option value="Streaming">Streaming</option>
                <option value="Recording">Recording</option>
            </select>
            <br/>
            <label for="green_value">Tally Value</label>
            <input type="text" name="green_value" id="green_value" size=3/>
            <br/>
            <label for="ssid">WiFi SSID</label>
            <input type="text" name="ssid" id="ssid"/>
            <br/>
            <label for="wifi_password">WiFi Password</label>
            <input type="text" name="wifi_password" id="wifi_password"/>
            <br/>
            <label for="api_ip">vMix API IP address</label>
            <input type="text" name="api_ip" id="api_ip"/>
            <br/>
            <br/>
            <input type="submit" name="sub_form" id="sub_form" value="Submit"/>
        </form>
     </div>
 </div>
</body>
</html>
<?php } ?>