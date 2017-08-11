#!/bin/sh

ACK="ack"
SESSION="test"

function send()
{
    message=$1

    for port in {4001..4004}
    do
	ack=`echo $message | netcat "localhost" $port`
	if [ "$ack" != "$ACK" ]
	then
	    echo $message" failed"
	    exit 1
	fi
    done
}

# init
go build
./mikser -silent -host=localhost:4001 &
./mikser -silent -host=localhost:4002 &
./mikser -silent -host=localhost:4003 &
./mikser -silent -host=localhost:4004 &

# distributed key generation
send "start_dkg\n"$SESSION
echo "DKG setup completed" && sleep 2s
send "start_deal\n"$SESSION
echo "Deals distributed" && sleep 2s
send "start_response\n"$SESSION
echo "Responses distributed" && sleep 2s
send "start_commit\n"$SESSION
echo "Certification successful" && sleep 2s

# shared key retrieval
message="shared_key\n"$SESSION
echo "Key 1: "`echo $message | netcat "localhost" 4001`
echo "Key 2: "`echo $message | netcat "localhost" 4002`
echo "Key 3: "`echo $message | netcat "localhost" 4003`
echo "Key 4: "`echo $message | netcat "localhost" 4004`
