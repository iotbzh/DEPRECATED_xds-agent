#!/bin/bash

# Install XDS agent as a user systemd service

DESTDIR=${DESTDIR:-/opt/AGL/xds/agent}
DESTDIR_WWW=${DESTDIR_WWW:-${DESTDIR}/www}

ROOT_SRCDIR=$(cd $(dirname "$0")/.. && pwd)

install() {
    mkdir -p ${DESTDIR} && cp ${ROOT_SRCDIR}/bin/* ${DESTDIR} || exit 1
    mkdir -p ${DESTDIR_WWW} && cp -a ${ROOT_SRCDIR}/webapp/dist/* ${DESTDIR_WWW} || exit 1

    cp ${ROOT_SRCDIR}/conf.d/etc/xds-agent /etc/ || exit 1
    cp ${ROOT_SRCDIR}/conf.d/etc/default/xds-agent /etc/default/ || exit 1

    FILE=/etc/profile.d/xds-agent.sh
    sed -e "s;%%XDS_INSTALL_BIN_DIR%%;${DESTDIR};g" ${ROOT_SRCDIR}/conf.d/${FILE} > ${FILE} || exit 1

    FILE=/usr/lib/systemd/user/xds-agent.service
    sed -e "s;/opt/AGL/xds/agent;${DESTDIR};g" ${ROOT_SRCDIR}/conf.d/${FILE} > ${FILE} || exit 1

    echo ""
    echo "To enable xds-agent service, execute:      systemctl --user enable xds-agent"
    echo "and to start xds-agent service, execute:   systemctl --user start xds-agent"
}

uninstall() {
    rm -rf "${DESTDIR}"
    rm -f /etc/xds-agent /etc/profile.d/xds-agent.sh /usr/lib/systemd/user/xds-agent.service
}

if [ "$1" == "uninstall" ]; then
    echo -n "Are-you sure you want to remove ${DESTDIR} [y/n]? "
    read answer
    if [ "${answer}" = "y" ]; then
        uninstall
        echo "xds-agent sucessfully uninstalled."
    else
        echo "Uninstall canceled."
    fi
else
    install
fi
