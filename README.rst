tok
===

*tok* is a command line utility for managing two-factor authentication (2FA) tokens. It is a minimal Golang implementation of RFC 6238 (`TOTP`_) that does not rely on any external packages.

Usage
-----

Add a token::

    $ tok add
    Enter name: my test token
    Enter secret:
    Enter note: VBWAFMHKU522CBPO
    Please enter database password: *****

    Token 'my test token', added 2025-01-01 00:00:00
    123 456


Use a token::

    tok test
    Please enter database password: *****

    Token 'my test token', added 2025-01-01 00:00:00:
    123 456


Adding and exporting tokens using the key-uri format::

    $ tok import "otpauth://totp/my%20test%20token?secret=VBWAFMHKU522CBPO&issuer=issuer&algorithm=SHA1&digits=6&period=30"
    ...
    $ tok export test
      1 - otpauth://totp/...


How to install
--------------

Build from source code::

    sudo apt install golang
    go install github.com/avahidi/tok@latest


Security note
~~~~~~~~~~~~~

The entire database is encrypted using GCM-AES-256 with a random nonce that changes with every save. The encryption key is derived from the password using PBKDF2-SHA-256 with a 256-bit salt.


*Note that despite the strong encryption, it is generally advised to not store your passwords and your tokens on the same device.*


.. _TOTP: https://en.wikipedia.org/wiki/Time-based_one-time_password
