#!/bin/bash

set -e

CLOUDHSM_IP=${CLOUDHSM_IP:-127.0.0.1}
HSM=${HSM:-cloudhsm}

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
  ;;
*)
  echo "unkonwn command: ${HSM}"
  exit 1
esac

if [ ! -f /opt/cloudhsm/etc/customerCA.crt ]; then
  echo "warning: /opt/cloudhsm/etc/customerCA.crt not found"
fi

/tm-pkcs11 "$@"
