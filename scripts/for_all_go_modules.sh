#!/bin/bash

# Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

export GIT_ROOT="$(git rev-parse --show-toplevel)"

cd $GIT_ROOT

EXCLUDES=""
PARALLEL=8
REQUIRED_FILES=""
ADDITIONAL_TARGETS=""

while [ "$1" != "--" ]; do
  if [ "$1" == "--exclude" ]; then
    shift
    EXCLUDES="$EXCLUDES -e $1"
  fi

  if [ "$1" == "--parallel" ]; then
    shift
    PARALLEL="$1"
  fi

  if [ "$1" == "--with-file" ]; then
    shift
    REQUIRED_FILES="$REQUIRED_FILES $1"
  fi

  if [ "$1" == "--add-target" ]; then
    shift
    ADDITIONAL_TARGETS="$ADDITIONAL_TARGETS $1"
  fi

  shift
done

export PARALLEL
export EXCLUDES
export REQUIRED_FILES
export ADDITIONAL_TARGETS

shift

collectAllModules() {
  echo .
  find */* -name go.mod  | sed 's?/go.mod$??'
  if [ -n "${ADDITIONAL_TARGETS}" ]; then
    for T in ${ADDITIONAL_TARGETS}; do
      echo $T
    done
  fi
}

filterExcludes() {
  if [ -z "$EXCLUDES" ]; then
    cat
  else
    grep -v $EXCLUDES
  fi
}

filterDirsWithRequiredFiles() {
  while read CANDIDATE_FOLDER; do
    if [ -z "$REQUIRED_FILES" ]; then
      echo $CANDIDATE_FOLDER
    else
      MISSING_FILES=0
      for RQ_FILE in $REQUIRED_FILES; do
        if [ ! -f $CANDIDATE_FOLDER/$RQ_FILE ]; then
          MISSING_FILES=1
        fi
      done

      if [ "$MISSING_FILES" == 0 ]; then
        echo $CANDIDATE_FOLDER
      fi
    fi
  done
}

execute() {

  if [ "${PARALLEL}" == 1 ]; then
    while read DIR; do
      ${GIT_ROOT}/scripts/execute_command.sh $@ $DIR
    done
  else
    xargs -P ${PARALLEL} -n 1 ${GIT_ROOT}/scripts/execute_command.sh $@
  fi
}

collectAllModules | filterExcludes | filterDirsWithRequiredFiles | execute $@
