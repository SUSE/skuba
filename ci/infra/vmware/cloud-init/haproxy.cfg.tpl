global 
  log /dev/log local0 debug
  user haproxy
  group haproxy
  daemon

defaults
  mode      tcp
  log       global
  option    tcplog
  option    redispatch
  option    tcpka
  retries   2
  http-check     expect status 200
  default-server check check-ssl verify none
  timeout connect 5s
  timeout client 5s
  timeout server 5s
  timeout tunnel 86400s

listen stats
  bind    *:9000
  mode    http
  stats   hide-version
  stats   uri       /stats

listen apiserver
  bind    *:6443
  option  httpchk GET /healthz
  ${apiserver_backends}

listen dex
  bind    *:32000
  option  httpchk GET /healthz
  ${dex_backends}

listen gangway
  bind    *:32001
  option httpchk GET /
  ${gangway_backends}
