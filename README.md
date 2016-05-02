Stress Tester
=============

This is a MQTT broker stress tester designed to determine the capacity of a
broker by bombarding it with concurrent messages.

This program is written in the Go programming language and was developed and
tested with Go 1.6.

Building
--------

Build this program with the command

```
make all
```

Testing
-------

You can run the unit-tests with the command

```
make test
```

Running
-------

The command-line usage can be requested with the ``--help`` or ``-h`` flags:

```
bash$ mqtt_stresser -h
Usage:
  mqtt_stresser [OPTIONS]

Broker Connection Options:
      --hostname=            Address of the broker to connect to (default:
                             localhost)
      --passwd-file=         File with raw-text usernames and passwords
  -u, --username=            Name of the user to connect with. Superceded by
                             --passwd-file if specified
  -P, --password=            Password of the user to connect with. Used in
                             tandem with username
  -p, --port=                The port to connect through (default: 1883)

Publish/Subscribe Options:
  -n, --num-publishers=      Number of concurrent publishers (default: 1)
  -m, --messages-per-second= Average number of messages per second to send from
                             each publisher (default: 10)
  -d, --duration=            Number of seconds to run the publishers for
                             (default: 5)
  -s, --message-size=        Average number of bytes per message. At least 8
                             needed to collect timing data (default: 50)
  -v, --msg-rate-variance=   Variance (seconds squared) of the sample of
                             message rates (default: 0.005)
  -V, --msg-size-variance=   Variance (messages squared) of the sample of
                             message sizes (default: 5)
  -t, --topic-prefix=        Prefix to add to all random topic names for each
                             publisher (default: test/)

Input/Output Files:
  -c, --ca-file=             Certificate authority to enable anonymous TLS
                             connection
  -o, --output=              Output file to write detailed pub/sub statistics
                             to (default: stdout)
  -y, --yaml=                Input file with command-line parameters in YAML
                             format. CL options appearing before are
                             overridden. Those appearing after override.

Help Options:
  -h, --help                 Show this help message
```

The command-line options afford you the flexibility to publish more or fewer
messages, at whatever rate you choose.

Alternative Input: YAML
~~~~~~~~~~~~~~~~~~~~~~~

If you do not want to specify all input options on the command-line, you can
write a YAML-formatted input file whose keys are the same as the long
command-line options shown above. An example file is shown below:

```yaml
# Sample YAML file for input to mqtt_stresser. The YAML names are the same as
# the command-line flags (without the leading -- or short options)
hostname: somehost
passwd-file: passwd
username: username
password: password
port: 1985
num-publishers: 100
messages-per-second: 100
msg-rate-variance: 1.0
message-size: 100
msg-size-variance: 10
duration: 120
topic-prefix: sometopic/
ca-file: some.crt
output: some.csv
```

The way input is processed is that YAML files are dispatched for processing as
soon as they are found on the command-line. Because of this order of operations,
any command-line argument appearing *before* the YAML file will be overwritten,
while any argument appearing *after* the YAML file will override the contents of
the YAML file.
