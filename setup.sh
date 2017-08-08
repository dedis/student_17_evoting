message="generate\nhello"

echo $message | netcat "localhost" "4001" 
echo $message | netcat "localhost" "4002"
echo $message | netcat "localhost" "4003"
echo $message | netcat "localhost" "4004"

sleep 5s

cert="cert\nhello"
echo $cert | netcat "localhost" "4001" 
echo $cert | netcat "localhost" "4002"
echo $cert | netcat "localhost" "4003"
echo $cert | netcat "localhost" "4004"

sleep 2s

key="key\nhello"
echo $key | netcat "localhost" "4001"
echo $key | netcat "localhost" "4002"
echo $key | netcat "localhost" "4003"
echo $key | netcat "localhost" "4004"
