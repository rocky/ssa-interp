#!/usr/bin/env bash

if (( $# < 1 )) ; then
    cat <<EOF
usage:
   $0 <go-program> [program-options]

Runs Go SSA debugger
EOF
    exit 1
fi

# Find tortoise
tortoise=$(which tortoise)
if (( $? != 0 )); then
    dirname=${BASH_SOURCE[0]%/*}
    tortoise="$dirname/tortoise"
    [[ ! -x $tortoise ]] && {
	builtin echo "Can't find tortoise in PATH or as $tortoise";
	exit 2
    }
fi

# Run tortoise setting
$tortoise -run -interp=S -- $@
exit $?
