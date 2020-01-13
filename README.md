#  Vault storage (ezb_vault)

The vault service, store key/value pair in a central store. It's used to not hardcode data in worker's scripts, like password or constant.

## Use case

### New secret
- Get your JWT auth token from a ezb_sta.
- Create header and body.
- Call the vault.
```powershell
$a = Invoke-RestMethod -Uri http://ezb_sta.fqdn/token -UseDefaultCredentials
if($a) {
  $h = @{}
  $h.Authorization = "bearer "+ $a.access_token
  $h."EZB-VAULT-KEY" = "AEScryptKEY"
  $key = @{}
  $key.key = "firstkey"
  $key.value = "firstvalue"
  Invoke-RestMethod -Headers $h -Uri https://ezb_vault.fqdn -Method Post -Body $( $key | ConvertTo-Json -Compress) -ContentType "application/json"
}
```

### Retrieve secret
- one

```powershell
Invoke-RestMethod -Headers $h -Uri https://ezb_vault.fqdn/firstkey
```

- all

```powershell
Invoke-RestMethod -Headers $h -Uri https://ezb_vault.fqdn
```
### Update a secret
```powershell
Invoke-RestMethod -Headers $h -Uri https://ezb_vault.fqdn/firstkey -Method Put -Body $( $key | ConvertTo-Json -Compress) -ContentType "application/json"
```

### Delete a secret
```powershell
Invoke-RestMethod -Headers $h -Uri https://ezb_vault.fqdn/firstkey -Method Delete
```

## SETUP


### 1. Download ezb_vault from [GitHub](<https://github.com/ezBastion/ezb_vault/releases/latest>)

### 2. Open an admin command prompte, like CMD or Powershell.

### 3. Run ezb_vault.exe with **init** option.

```powershell
    PS E:\ezbastion\ezb_vault> ezb_vault init
```

this commande will create folder and the default config.json file.
```json
{
    "listen": ":5100",
    "privatekey": "cert/ezb_vault.key",
    "publiccert": "cert/ezb_vault.crt",
    "cacert": "cert/ca.crt",
    "dbpath": "db/ezb_vault.db",
    "servicename": "ezb_vault",
    "servicefullname": "Easy Bastion Vault",
    "loglevel": "warning"
}
```
> /!\ Don't forget to copy all public STA certificat to the cert folder /!\
> cert name must match jwt ISS value.



### 4. Install Windows service and start it.

```powershell
    PS E:\ezbastion\ezb_vault> ezb_vault install
    PS E:\ezbastion\ezb_vault> ezb_vault start
```




## Copyright

Copyright (C) 2018 Renaud DEVERS info@ezbastion.com
<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL%20v3-blueviolet.svg?style=for-the-badge&logo=gnu" alt="License"></a></p>


Used library:

Name      | Copyright | version | url
----------|-----------|--------:|----------------------------
gin       | MIT       | 1.2     | github.com/gin-gonic/gin
cli       | MIT       | 1.20.0  | github.com/urfave/cli
gorm      | MIT       | 1.9.2   | github.com/jinzhu/gorm
logrus    | MIT       | 1.0.4   | github.com/sirupsen/logrus
go-fqdn   | Apache v2 | 0       | github.com/ShowMax/go-fqdn
jwt-go    | MIT       | 3.2.0   | github.com/dgrijalva/jwt-go
gopsutil  | BSD       | 2.15.01 | github.com/shirou/gopsutil

