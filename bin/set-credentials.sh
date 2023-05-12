#!/usr/bin/env bash

#
# Copyright (C) 2023 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to allow end-users to set credentials either through
# EdgeX InsecureSecrets via Consul, or EdgeX Secrets via the device service.
#

set -euo pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]:-${0}}")")"

# shellcheck source=./utils.sh
source "${SCRIPT_DIR}/utils.sh"

main() {
    parse_args "$@"

    dependencies_check

    # Only non-secure mode needs to use Consul. Secure mode only uses device service APIs
    if [ "${SECURE_MODE}" -ne 1 ]; then
      consul_check
    fi

    if  [ "${SECURE_MODE}" -eq 1 ] && [ -z "${REST_API_JWT}" ]; then
        query_rest_api_jwt
    fi

    query_username_password
    set_secret

    echo -e "${green}${bold}Success${clear}"
}

main "$@"
