#!/bin/bash

# Copyright 2021, CTERA Networks.
#
# Portions Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -euo pipefail


if [ $# -lt 2 ]; then
    echo "Must pass report file and expected coverage"
    exit 1
fi

REPORT_FILE=$1
EXPECTED_COVERAGE=$2

if [ ! -f "${REPORT_FILE}"  ]; then
    echo "File ${REPORT_FILE} does not exist"
    exit 1
fi

if ! [[ "${EXPECTED_COVERAGE}" =~ ^[0-9]+$ ]] ; then
    echo "Expected coverage must be an integer"
    exit 1
fi

readonly coverage=$(go tool cover -func=${REPORT_FILE} | grep total | awk '{print $3}' | awk -F '.' '{print $1}')
if [ ${coverage} -lt ${EXPECTED_COVERAGE} ]; then
    echo "Insufficient coverage"
    exit 1
fi
