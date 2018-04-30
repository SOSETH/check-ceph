# check-ceph

Check CEPH cluster health using CEPH's json status output. This has been tested
with CEPH releases >= Luminous.

## How to build

```
# go build
```
... that's it. There are no dependencies except the standard library.

## How to use with Icinga2

You'll need a command definition:
```
object CheckCommand "ceph {
  import "plugin-check-command"
  command = [PluginDir + "/check-ceph" ]
}
```
...and a matching template...
```
apply Service "ceph" {
  import "normal-service"
  check_command = "ceph"
  assign where host.vars.ceph
}
```
and finally tag the respective hosts with:
```
  vars.ceph = "true"
```
