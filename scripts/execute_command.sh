#!/bin/bash

set -euo pipefail

DIR=${!#}
set -- "${@:1:$#-1}"


cd $DIR

echo "[$DIR] Execute: $@"

(
  $@ 2>&1
) | awk "{ print \"[$DIR] \" \$0 }"
