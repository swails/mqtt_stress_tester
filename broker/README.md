MQTT Broker
===========

The files here are configuration and password files for setting up an MQTT
broker on ``localhost`` in order to support a variety of tests testing MQTT
communications.

Testing TLS
-----------

Testing TLS encryption in this code base relies on using ``openssl`` to generate
self-signed certificates to authenticate the communication. Because the private
key is provided here, the certificates in this test directory are for testing
purposes only.

IMPORTANT: DO NOT USE THE SSL CERTIFICATES IN THIS REPOSITORY FOR ANY PRODUCTION
WORK. THEY ARE FOR THE TESTS ONLY! These certificates are very insecure since
*the private key is now public*.

These commands were used to generate the certificate authority as well as the
broker certificates and certificate signing request:

```bash
# Create the certificate authority. Use password "stresser" and confirm
openssl req -new -x509 -extensions v3_ca -keyout ca.key -out ca.crt -days 3650 -config openssl.cnf -subj '/CN=TEST CA/'
# Create the private key for the broker
openssl genrsa -out broker.key 2048 -config openssl.cnf
# Create a certificate signing request from the broker private key
openssl req -out broker.csr -key broker.key -new -config openssl.cnf -subj '/CN=localhost/O=Stresser Tester, Inc./C=US'
# Sign the CSR with the certificate authority. You will need the CA password you used in step 1
openssl x509 -req -in broker.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out broker.crt -days 3650
```

Starting the broker
-------------------

The configuration files here are designed to run ``mosquitto`` as an MQTT broker
on ``localhost``.

Special note
------------

Make sure you copy the root certificate (``ca.crt``) to the ``mqtt`` package so
the tests can find it.

Users and Passwords
-------------------

To emulate real-world scenarios with large-scale brokers, a ``mosquitto.passwd``
file has been generated with one pre-defined username/password combination
(``stresstest`` user and ``stressmeout`` password). Another 1000 random
usernames and passwords have been generated as well by ``random_userpass.py``.
The raw text users and passwords are stored in ``rawtext.passwd``, and the
command

```
mosquitto_passwd -U mosquitto.passwd
```

was run on a copy of this raw text password file to generate a file with user
names and hashed passwords. This allows the stress tester to read a raw text
file with passwords and user accounts to make multiple authenticated connections
with different accounts.
