#!/bin/bash

set -e

CLOUDHSM_IP=${CLOUDHSM_IP:-127.0.0.1}
HSM=${HSM:-softhsm}

function start_cloudhsm() {
  # start cloudhsm client
  echo -n "* Starting CloudHSM client ... "
  /opt/cloudhsm/bin/cloudhsm_client /opt/cloudhsm/etc/cloudhsm_client.cfg &> /tmp/cloudhsm_client_start.log &

  # wait for startup
  while true
  do
      if grep 'libevmulti_init: Ready !' /tmp/cloudhsm_client_start.log &> /dev/null
      then
          echo "[OK]"
          break
      fi
      sleep 0.5
  done
  echo -e "\n* CloudHSM client started successfully ... \n"
}

case ${HSM} in
cloudhsm)
  /opt/cloudhsm/bin/configure -a "${CLOUDHSM_IP}"
  start_cloudhsm
  if [ ! -f /opt/cloudhsm/etc/customerCA.crt ]; then
    echo "warning: /opt/cloudhsm/etc/customerCA.crt not found"
  fi
  ;;
softhsm)
  ;;
*)
  echo "unkonwn command: ${HSM}"
  exit 1
esac

function handle_error() {
  echo "Failed to start tm-pkcs11 ($1)."
  case ${HSM} in
  cloudhsm)
    cat /tmp/cloudhsm_client_start.log
    ;;
  esac
  exit "$1"
}

/tm-pkcs11 "$@" || handle_error $?
