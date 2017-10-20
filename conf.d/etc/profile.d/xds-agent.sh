#!/bin/bash

#----------  AGL xds-agent tool options Start ---------"
[ ":${PATH}:" != *":%%XDS_INSTALL_BIN_DIR%%:"* ] && export PATH=%%XDS_INSTALL_BIN_DIR%%:${PATH}
