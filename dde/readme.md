Periodic Ping
=============

DDE Connector client will be disconnected after 3 minutes.  
After disconnected, client will auto restart to connect to DDE Connector server again.  

    2018-05-17 11:45:59.171 Restarting
    2018-05-17 11:48:59.599 Restarting
    2018-05-17 11:51:59.724 Restarting
    2018-05-17 11:55:00.099 Restarting
    
Periodic sending "> Ping" for every 60s avoid the TCP disconnection.
It is tested that after "> Ping" the TCP connection never restart.

London Time
===========

London time uses Daylight Saving Time. The program uses time.LoadLocation().
It requires IANA Time Zone database of below items

    Europe/London
    Asia/Hong_Kong
    
On Windows production, it requires to install Go so that below zip IANA Time Zone database exists

    $GOROOT/lib/time/zoneinfo.zip
    
Firewall
========

Universal DDE Connector uses TCP 2222
The file here C:\Program Files (x86)\UniDDEConnector\default.sym stores the 

Items and Decimal Points
========================

GOLD   2
SILVER 2
USDCHF 5
GBPUSD 5
EURUSD 5
USDJPY 3
USDCAD 5
AUDUSD 5
EURGBP 5
EURAUD 5
EURCHF 5
EURJPY 3
GBPCHF 5
CADJPY 3
GBPJPY 3
AUDNZD 5
AUDCAD 5
AUDCHF 5
AUDJPY 3
CHFJPY 3
EURNZD 5
EURCAD 5
CADCHF 5
NZDJPY 3
NZDUSD 5


MT4
===

Suggest using RBTray for minimizing MT4 terminal to tray on production server