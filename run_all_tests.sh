# Fire up the mosquitto broker
if [ -z "`which mosquitto 2>/dev/null`" ]; then
  echo "You need mosquitto to run these tests."
  exit 1
fi

GOPATH="$PWD"/"`dirname $0`"

echo "Starting broker"
cd broker && mosquitto -c mosquitto.conf &
sleep 1 # Let the broker start up

cd ../

echo "Setting GOPATH to $GOPATH"
export GOPATH

echo "Running Go tests"

go test -v mqtt/... messages/... -timeout 5s

rc=$?

echo "Killing the mosquitto broker"
killall mosquitto

if [ $rc -ne 0 ]; then
  echo "Test failures detected!"
else
  echo "Tests succeeded!"
fi
exit $rc
