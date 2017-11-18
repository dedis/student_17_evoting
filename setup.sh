#!/bin/bash

set -e

CONODE=nevv
GOPATH=${GOPATH:-`go env GOPATH`}
CONODE_GO=github.com/qantik/nevv

main() {
    if [ ! "$GOPATH" ]; then
    	echo "'$GOPATH' not found"
    	exit 1
    fi
    
    if ! echo $PATH | grep -q "$GOPATH/bin"; then
    	PATH=$PATH:$GOPATH/bin
    fi

    if [ "$#" -lt 1 ]; then
        echo "Specify task"
        exit 1
    fi
    
    case $1 in
        run)
    	    run $@ ;;
        stop)
    	    stop ;;
        test)
    	    test ;;
        clean)
    	    clean ;;
        *)
    	    echo "Task '$ACTION' not found." ;;
    esac
}

run() {
    if [ ! "$2" ]; then
    	echo "Specify number of nodes"
    	exit 1
    fi
    
    NODES=$2
    DEBUG=1
    
    if [ "$3" ]; then
        DEBUG=$3
    fi
    
    killall -9 $CONODE 2> /dev/null || true
    go install $CONODE_GO
    rm -f public.toml
    
    for n in $( seq $NODES ); do
    	co=co$n
    	if [ -f $co/public.toml ]; then
	    if ! grep -q Description $co/public.toml; then
    		rm -rf $co
	    fi
    	fi
    
    	if [ ! -d $co ]; then
    		echo -e "127.0.0.1:$((7000 + 2 * $n))\nConode_$n\n$co" | $CONODE setup
    	fi
    
    	$CONODE -c $co/private.toml -d $DEBUG &
    	cat $co/public.toml >> public.toml
    done
    
    sleep 1
    echo "Conodes setup successful. Use public.toml to interact with the cothority."
}

stop() {
    killall -9 $CONODE 2> /dev/null || true
}

test() {
    go test ./...
}

clean() {
    rm -rf co* public.toml 2> /dev/null
}

main $@
