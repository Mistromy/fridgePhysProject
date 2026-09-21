Shelly works

endpoint: http://192.168.33.1/rpc/Switch.GetStatus?id=0
api shape:
```json
{
  "id":0,
  "source":"WS_in", 
  "output":true, 
  "apower":4.1, 
  "voltage":242.5, 
  "current":0.033, 
  "aenergy":{"total":18.015},
  "temperature":{"tC":30.3, "tF":86.5}
}
```
`(182B)`

`182B` * `60` * `60`  = `655kB h⁻¹`

settings:
http://192.168.33.1/rpc/Switch.GetConfig?id=0
```json
{
  "id":0, 
  "name":null,
  "initial_state":"on", 
  "auto_on":false, 
  "auto_on_delay":60.00, 
  "auto_off":false, 
  "auto_off_delay":60.00,
  "power_limit":3000,
  "voltage_limit":280,
  "autorecover_voltage_errors":false,
  "current_limit":13.000
}
```
changed settings:
http://192.168.33.1/rpc/Switch.SetConfig?id=0&config=%7B%22initial_state%22:%22on%22%7D
```json
{"restart_required":false}
```