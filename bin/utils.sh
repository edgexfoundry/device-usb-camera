#!/usr/bin/env bash

#
# Copyright (C) 2022-2023 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0
#

#
# The purpose of this script is to be sourced by other utility scripts from
# this service in order to reduce duplicated code.
#

CONSUL_URL="${CONSUL_URL:-http://localhost:8500}"
DEVICE_SERVICE="${DEVICE_SERVICE:-device-usb-camera}"
DEVICE_SERVICE_URL="${DEVICE_SERVICE_URL:-http://localhost:59983}"

SECRET_NAME="${SECRET_NAME:-rtspauth}"
SECRET_USERNAME="${SECRET_USERNAME:-}"
SECRET_PASSWORD="${SECRET_PASSWORD:-}"

CONSUL_KV_BASE_URL="${CONSUL_URL}/v1/kv"
CONSUL_BASE_KEY="edgex/v3/${DEVICE_SERVICE}"
WRITABLE_BASE_KEY="${CONSUL_BASE_KEY}/Writable"
INSECURE_SECRETS_KEY="${WRITABLE_BASE_KEY}/InsecureSecrets"

CONSUL_TOKEN="${CONSUL_TOKEN:-}"
REST_API_JWT="${REST_API_JWT:-}"

CURL_CODE=
CURL_OUTPUT=

SECURE_MODE=${SECURE_MODE:-0}

SELF_CMD="${0##*/}"

# ANSI colors
red="\033[31m"
green="\033[32m"
clear="\033[0m"
bold="\033[1m"
dim="\033[2m"
normal="\e[22;24m"

# these are used for printing out messages
spacing=18
prev_line="\e[1A\e[$((spacing + 2))C"

# print a message in bold
log_info() {
    echo -e "${bold}$*${clear}"
}

# print a message dimmed
log_debug() {
    echo -e "${dim}$*${clear}"
}

# log an error message to stderr in bold and red
log_error() {
    echo -e "${red}${bold}$*${clear}" >&2
}

# attempt to pretty print the output with jq. if jq is not available or
# jq fails to parse data, print it normally
format_output() {
    if [ ! -x "$(type -P jq)" ] || ! jq . <<< "$1" 2>/dev/null; then
        echo "$1"
    fi
    echo
}

# call the curl command with the specified payload and arguments.
# this function will print out the curl response and will return an error code
# if the curl request failed.
# usage: do_curl "<payload>" curl_args...
do_curl() {
    local payload="$1"
    shift

    # log the curl command so the user has insight into what the script is doing
    # redact the consul token and password in case of sensitive data
    local redacted_args="$*"
    redacted_args="${redacted_args//${CONSUL_TOKEN}/<redacted>}"
    redacted_args="${redacted_args//${REST_API_JWT}/<redacted>}"
    local redacted_data=""
    if [ -n "${payload}" ]; then
        redacted_data="--data '${payload//${SECRET_PASSWORD}/<redacted>}' "
    fi
    log_debug "curl ${redacted_data}${redacted_args}" >&2

    local tmp code output
    # the payload is securely transferred through an auto-closing named pipe.
    # this prevents any passwords or sensitive data being on the command line.
    # the http response code is written to stdout and stored in the variable 'code', while the full http response
    # is written to the temp file, and then read into the 'output' variable.
    tmp="$(mktemp)"
    code="$(curl -sS --location -w "%{http_code}" -o "${tmp}" "$@" --data "@"<( set +x; echo -n "${payload}" ) || echo $?)"
    output="$(<"${tmp}")"

    declare -g CURL_CODE="$((code))"
    declare -g CURL_OUTPUT="${output}"
    printf "Response [%3d] " "$((code))" >&2
    if [ $((code)) -lt 200 ] || [ $((code)) -gt 299 ]; then
        format_output "$output" >&2
        log_error "Failed! curl returned a status code of '$((code))'"
        return $((code))
    else
        format_output "$output"
    fi

    echo >&2
}

query_rest_api_jwt() {
    REST_API_JWT=$(whiptail --inputbox "Enter REST API JWT (make get-token)" \
        10 0 3>&1 1>&2 2>&3)

    if [ -z "${REST_API_JWT}" ]; then
        log_error "No REST API JWT entered, exiting..."
        return 1
    fi
}

# usage: get_consul_kv <path from base> <args>
get_consul_kv() {
    do_curl "" -H "X-Consul-Token:${CONSUL_TOKEN}" -X GET "${CONSUL_KV_BASE_URL}/$1?$2"
}

# usage: put_consul_kv <key path from base> <value>
put_consul_kv() {
    do_curl "$2" -H "X-Consul-Token:${CONSUL_TOKEN}" -X PUT "${CONSUL_KV_BASE_URL}/$1"
}

# set an individual InsecureSecrets consul key to a specific value
# usage: put_insecure_secrets_field <sub-path> <value>
put_insecure_secrets_field() {
    log_info "Setting InsecureSecret: $1"
    put_consul_kv "${INSECURE_SECRETS_KEY}/$1" "$2"
}

# prompt the user for the credential's username and password
# and exit if not provided
query_username_password() {
    if [ -z "${SECRET_USERNAME}" ]; then
        SECRET_USERNAME=$(whiptail --inputbox "Enter username for ${SECRET_NAME}" \
            10 0 3>&1 1>&2 2>&3)

        if [ -z "${SECRET_USERNAME}" ]; then
            log_error "No username entered, exiting..."
            return 1
        fi
    fi

    if [ -z "${SECRET_PASSWORD}" ]; then
        SECRET_PASSWORD=$(whiptail --passwordbox "Enter password for ${SECRET_NAME}" \
            10 0 3>&1 1>&2 2>&3)

        if [ -z "${SECRET_PASSWORD}" ]; then
            log_error "No password entered, exiting..."
            return 1
        fi
    fi
}

# usage: try_set_argument "arg_name" "$@"
# attempts to set the global variable "arg_name" to the next value from the command line.
# if one is not provided, print error and return and error code.
# note: call shift AFTER this, as we want to see the flag_name as first argument after arg_name
try_set_argument() {
    local arg_name="$1"
    local flag_name="$2"
    shift 2
    if [ "$#" -lt 1 ]; then
        log_error "Missing required argument: ${flag_name} ${arg_name}"
        return 2
    fi
    declare -g "${arg_name}"="$1"
}

print_usage() {
    log_info "Usage: ${SELF_CMD} [-s/--secure-mode] [-u <username>] [-p <password>] [-t <consul token>]"
}

parse_args() {
    while [ "$#" -gt 0 ]; do
        case "$1" in

        -s | --secure | --secure-mode)
            SECURE_MODE=1
            ;;

        -t | --token | --consul-token)
            try_set_argument "CONSUL_TOKEN" "$@"
            shift
            ;;

        -u | --user | --username)
            try_set_argument "SECRET_USERNAME" "$@"
            shift
            ;;

        -p | --pass | --password)
            try_set_argument "SECRET_PASSWORD" "$@"
            shift
            ;;

        -c | --consul-url)
            try_set_argument "CONSUL_URL" "$@"
            shift
            ;;

        -m | --core-metadata-url)
            try_set_argument "CORE_METADATA_URL" "$@"
            shift
            ;;

        -U | --device-service-url)
            try_set_argument "DEVICE_SERVICE_URL" "$@"
            shift
            ;;

        --help)
            print_usage
            exit 0
            ;;

        *)
            log_error "argument \"$1\" not recognized."
            return 1
            ;;

        esac

        shift
    done
}

# create or update the insecure secrets by setting the 3 required fields in Consul
set_insecure_secret() {
    put_insecure_secrets_field "${SECRET_NAME}/SecretName"             "${SECRET_NAME}"
    put_insecure_secrets_field "${SECRET_NAME}/SecretData/username"    "${SECRET_USERNAME}"
    put_insecure_secrets_field "${SECRET_NAME}/SecretData/password"    "${SECRET_PASSWORD}"
}

# set the secure secrets by posting to the device service's secret endpoint
set_secure_secret() {
    if [ -z "${REST_API_JWT}" ]; then
        query_rest_api_jwt
    fi

    local payload="{
    \"apiVersion\":\"v3\",
    \"secretName\": \"${SECRET_NAME}\",
    \"secretData\":[
        {
            \"key\":\"username\",
            \"value\":\"${SECRET_USERNAME}\"
        },
        {
            \"key\":\"password\",
            \"value\":\"${SECRET_PASSWORD}\"
        }
    ]
}"
    do_curl "${payload}" \
        -H "Authorization:Bearer ${REST_API_JWT}" \
        -X POST "${DEVICE_SERVICE_URL}/api/v3/secret"
}

# helper function to set the secrets using either secure or insecure mode
set_secret() {
    if [ "${SECURE_MODE}" -eq 1 ]; then
        set_secure_secret
    else
        set_insecure_secret
    fi
}

# Dependencies Check
dependencies_check() {
    printf "${bold}%${spacing}s${clear}: ...\n" "Dependencies Check"
    if ! type -P curl >/dev/null; then
        log_error "${prev_line}${bold}${red}Failed!${normal}\nPlease install ${bold}curl${normal} in order to use this script!${clear}"
        return 1
    fi
    echo -e "${prev_line}${green}Success${clear}"
}

check_consul_return_code() {
    if [ $((CURL_CODE)) -ne 200 ]; then
        if [ $((CURL_CODE)) -eq 7 ]; then
            # Special message for error code 7
            echo -e "${red}* Error code '7' denotes 'Failed to connect to host or proxy'${clear}"
        elif [ $((CURL_CODE)) -eq 404 ]; then
            # Error 404 means it connected to consul but couldn't find the key
            echo -e "${red}* Have you deployed the ${bold}${DEVICE_SERVICE}${normal} service?${clear}"
        elif [ $((CURL_CODE)) -eq 401 ]; then
            if [ "${CURL_OUTPUT}" == "ACL support disabled" ]; then
                SECURE_MODE=0
                CONSUL_TOKEN=""
                return
            fi
            echo -e "${red}* Are you running in secure mode? Is your Consul token correct?${clear}"
        elif [ $((CURL_CODE)) -eq 403 ]; then
            # Error 401 and 403 are authentication errors
            if [ -z "${CONSUL_TOKEN}" ]; then
                SECURE_MODE=1
                # Note: USB device service does not need to use Consul at all for secure mode, so we do not need
                # to ask for the acl token
                return
            fi
            echo -e "${red}* Are you running in secure mode? Is your Consul token correct?${clear}"
        else
            echo -e "${red}* Is Consul deployed and accessible?${clear}"
        fi
        return $((CURL_CODE))
    fi
}

# Consul Check
consul_check() {
    printf "${bold}%${spacing}s${clear}: ...\n%${spacing}s  " "Consul Check" ""

    # use || true because we want to handle the result and not let the script auto exit
    do_curl '[{"Resource":"key","Access":"read"},{"Resource":"key","Access":"write"}]' \
        -H "X-Consul-Token:${CONSUL_TOKEN}" -X POST "${CONSUL_URL}/v1/internal/acl/authorize" 2>/dev/null || true
    check_consul_return_code


    if [ $((CURL_CODE)) -eq 200 ]; then
        local authorized
        # use || true because we want to handle the result and not let the script auto exit
        # this could be parsed better if using `jq`, but don't want to require the user to have it installed
        authorized=$(grep -c '"Allow":true'<<<"${CURL_OUTPUT}" || true)
        if [ $((authorized)) -ne 2 ]; then
            SECURE_MODE=1
            # Note: USB device service does not need to use Consul for secure mode, so no need to query acl token
            return
        fi
    fi

    # use || true because we want to handle the result and not let the script auto exit
    get_consul_kv "${CONSUL_BASE_KEY}" "keys=true" > /dev/null || true
    check_consul_return_code

    echo -e "${prev_line}${green}Success${clear}"
}
