# icalproxy

Sets up a webserver that proxies another iCal calendar, but modifies the response so that all events are considered all-day.

## Why?
D2L Brightspace's normal calendar link display events weird in the iOS Calendar app. I want them to be shown at the top of each day, so that I can see them easier. This program solves that issue, because all events (assignments) are marked as taking the entire day. Additionally, a reminder is added to all events. Just something I threw together in like 10 minutes with Claude, but it works. 

## Deployment (FreeBSD)

### Prerequisites

- Go toolchain installed (build step) or a pre-built binary
- Caddy already installed and configured with `basicauth`
- Root/sudo access for `pw`, `sysrc`, `service`

### 1. Build

```sh
go build -o icalproxy main.go
```

### 2. Install binary

```sh
install -o root -g wheel -m 755 icalproxy /usr/local/bin/icalproxy
```

### 3. Create config file

`/usr/local/etc/icalproxy.conf`:

```sh
icalproxy_source_url="https://example.com/original.ics"
icalproxy_listen="127.0.0.1:8080"
```

Lock down permissions:

```sh
install -o root -g wheel -m 600 icalproxy.conf /usr/local/etc/icalproxy.conf
```

### 4. Create unprivileged user

```sh
pw useradd icalproxy -d /nonexistent -s /usr/sbin/nologin -c "icalproxy daemon"
```

### 5. Install rc.d script

Save as `/usr/local/etc/rc.d/icalproxy`:

```sh
#!/bin/sh
#
# PROVIDE: icalproxy
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="icalproxy"
rcvar="icalproxy_enable"

load_rc_config "$name"

: ${icalproxy_enable:="NO"}
: ${icalproxy_conf:="/usr/local/etc/icalproxy.conf"}
: ${icalproxy_user_account:="icalproxy"}
: ${icalproxy_pidfile:="/var/run/${name}.pid"}
: ${icalproxy_bin:="/usr/local/bin/icalproxy"}

pidfile="$icalproxy_pidfile"
command="/usr/sbin/daemon"

start_precmd="icalproxy_prestart"

icalproxy_prestart()
{
	if [ -f "$icalproxy_conf" ]; then
		. "$icalproxy_conf"
	else
		err 1 "missing config file: $icalproxy_conf"
	fi

	export ICAL_PROXY_SOURCE_URL="$icalproxy_source_url"
	export ICAL_PROXY_LISTEN="$icalproxy_listen"

	command_args="-f -p ${pidfile} -u ${icalproxy_user_account} ${icalproxy_bin}"
}

run_rc_command "$1"
```

```sh
chmod +x /usr/local/etc/rc.d/icalproxy
```

### 6. Enable and start

```sh
sysrc icalproxy_enable="YES"
service icalproxy start
```

Verify:

```sh
service icalproxy status
curl http://127.0.0.1:8080/calendar.ics
```

### 7. Configure Caddy

Add a reverse proxy with `basicauth` in your Caddyfile, e.g.: