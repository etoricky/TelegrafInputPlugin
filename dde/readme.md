Periodic Ping
=============

DDE Connector client will be disconnected after 3 minutes.  
After disconnected, client will auto restart to connect to DDE Connector server again.  

    2018-05-17 11:45:59.171 Restarting
    2018-05-17 11:48:59.599 Restarting
    2018-05-17 11:51:59.724 Restarting
    2018-05-17 11:55:00.099 Restarting
    
Periodic sending "> Ping" for every 60s avoid the TCP disconnection.

London Time
===========

London time uses Daylight Saving Time. The program uses time.LoadLocation().
It requires IANA Time Zone database of below items

    Europe/London
    Asia/Hong_Kong
    
On Windows production, it requires to install Go so that below zip IANA Time Zone database exists

    $GOROOT/lib/time/zoneinfo.zip
    
    