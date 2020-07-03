#!/usr/bin/env bash
set -e

${ntp_servers}
${name_servers}
${register_scc}
${register_rmt}
${register_suma}
${repositories}
${commands}
