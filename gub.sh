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

TEMP=$(getopt -o hi:g: --long gub:,interp:,help -- "$@")

if [ $? != 0 ] ; then echo "Terminating..." >&2 ; exit 1 ; fi

# Note the quotes around `$TEMP': they are essential!
eval set -- "$TEMP"

gub_opt=''
interp_opt='S'
while true ; do
	case "$1" in
	    --gub) gub_opt="$2" ; shift ;;
	    --interp) interp_opt="S$2" ; shift ;;
	    --help|h) cat <<EOF
Usage: $0 *gub-opts* [--] *go-program* [*program options]

Runs SSA interpreter on *go-program* and gub debugger

opts are:

  --gub="--options to gub"
  --interp="options to tortoise interpeter"
  --help|-h  this help
EOF
		exit 100 ;;
	    --) shift;  break ;;
	    *) shift ;;
	esac
	shift
done


# Run tortoise setting
$tortoise -run -gub="$gub_opt" -interp="S$interp_opt" -- $@
exit $?
