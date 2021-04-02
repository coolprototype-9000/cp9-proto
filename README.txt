=============
     cp9
=============

Integrated, browser native, operating-system-like
environment offering per-process namespaces, locality
transparency, and idiomatic file-like interfaces to all
system objects.

To run, navigate to https://coolprototype9000.herokuapp.com, which is
running an instance of coolprototype-9000/shell-web-app in dev mode.
Once there, build cp9psrv by issuing `make` and then run it. Connect
to localhost:6969 in the browser and enjoy.

Supported commands for the CLI are based on 9P directly:

```
Tversion 0 <version>
Tattach <new fid u want> <your username (not checked)>
Topen <fid> <numerical open mode, see nine/param.go>
Twalk <fid> <new fid> <optional names to walk thru>
Tcreate <fid to assign> <filename> <num. perms> <openmode>
Tread <open fid> <offset> <count to read>
Twrite <open fid> <offset> <string to write...>
Tclunk <fid to clunk>
Tremove <fid to remove>
Tstat <fid to stat>
```
